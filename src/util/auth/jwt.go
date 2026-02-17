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
	ErrUserIsNil          = errors.New("error is")
)

// TODO: check if custom claims are needed

type Claims struct {
	jwt.RegisteredClaims
	TokenUuid uuid.UUID `json:"uuid"`
}

const (
	_ACCESS_TOKEN_DURATION  = 5 * time.Minute
	_REFRESH_TOKEN_DURATION = 7 * 24 * time.Hour
)

// ParseToken will parse the auth header string and return a token with claims or error
func ParseToken(authHeader string) (*jwt.Token, *Claims, error) {
	// Parse token
	if len(authHeader) <= len("Bearer ") || authHeader[:len("Bearer ")] != "Bearer " {
		zap.S().Debugf("token: %s", authHeader)
		return nil, nil, ErrInvalidTokenFormat
	}
	tokenString := authHeader[len("Bearer "):]
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(config.Config.CacheServer.Key), nil
	})
	if err != nil {
		return nil, nil, err
	}

	return token, &claims, nil
}

// GenerateJwt return a jwt access token and refresh token or an error
func GenerateJwt() (string, string, error) {
	uuidPair := uuid.New()
	accessTokenClaims := &Claims{
		TokenUuid: uuidPair,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(_ACCESS_TOKEN_DURATION)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(config.Config.CacheServer.Key))
	if err != nil {
		zap.S().Errorf("Failed to generate access token err = %w", err)
		return "", "", err
	}

	refreshTokenClaims := &Claims{
		TokenUuid: uuidPair,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(_REFRESH_TOKEN_DURATION)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.Config.CacheServer.Key))
	if err != nil {
		zap.S().Errorf("Failed to generate refresh token err = %w", err)
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
