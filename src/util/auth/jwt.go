package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/killi1812/go-cache-server/config"
	"go.uber.org/zap"
)

var (
	ErrInvalidTokenFormat = errors.New("error invalid token format")
	ErrTokenNotValid      = errors.New("error token is not valid ")
	ErrTokenRevoked       = errors.New("error token is revoked")
)

// TODO: add suport for multiple keys so that caches don't use app keys
// Suport for createing jwt token on first run

type Claims struct {
	Id          uuid.UUID        `json:"id"`
	ExpiresAt   *jwt.NumericDate `json:"expiresOn"`
	CreatedOn   *jwt.NumericDate `json:"createdOn"`
	LastUsedOn  *jwt.NumericDate `json:"lastUsedOn"` // not implemented currently
	Permission  string           `json:"permission"` // not implemented currently
	IsRevoked   bool             `json:"isRevoked"`
	Description string           `json:"description"`
}

// Valid implements jwt.Claims.
func (c *Claims) Valid() error {
	if c.IsRevoked {
		return ErrTokenRevoked
	}

	token := jwt.RegisteredClaims{
		ID:        c.Id.String(),
		ExpiresAt: c.ExpiresAt,
		IssuedAt:  c.CreatedOn,
	}

	return token.Valid()
}

const (
	_API_TOKEN_DURATION = 365 * 24 * time.Hour
)

// ParseToken will parse  the auth header string and verify the token, returns a claims or error
func ParseToken(authHeader string) (*Claims, error) {
	tokenString, err := cutHeader(authHeader)
	if err != nil {
		return nil, err
	}

	claims := new(Claims)
	err = parseToken(claims, tokenString)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func parseToken(claims *Claims, tokenString string) error {
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(config.Config.CacheServer.Key), nil
	})
	if err != nil {
		zap.S().Errorf("Parse and validation Failed, err: %v, %w", err, err)

		return err
	}
	return nil
}

func cutHeader(authHeader string) (string, error) {
	// Parse token
	if len(authHeader) <= len("Bearer ") || authHeader[:len("Bearer ")] != "Bearer " {
		zap.S().Debugf("token: %s", authHeader)
		return "", ErrInvalidTokenFormat
	}

	tokenString := authHeader[len("Bearer "):]
	return tokenString, nil
}

// GenerateJwt return a jwt api token or an error
func GenerateJwt() (string, error) {
	return GenerateJwtWithDuration(_API_TOKEN_DURATION)
}

// GenerateJwt return a jwt api token or an error
func GenerateJwtWithDuration(duration time.Duration) (string, error) {
	apiTokenClaims := &Claims{
		Id:        uuid.New(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		CreatedOn: jwt.NewNumericDate(time.Now()),
	}
	apiToken := jwt.NewWithClaims(jwt.SigningMethodHS256, apiTokenClaims)

	apiTokenString, err := apiToken.SignedString([]byte(config.Config.CacheServer.Key))
	if err != nil {
		zap.S().Errorf("Failed to generate api token err = %w", err)
		return "", err
	}

	return apiTokenString, nil
}
