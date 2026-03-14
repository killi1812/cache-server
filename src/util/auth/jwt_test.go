package auth_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/util/auth"
)

func TestParseToken(t *testing.T) {
	config.LoadConfig()

	now := time.Now()
	accessTokenDuration := 5 * time.Minute
	validClaims := auth.Claims{
		ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenDuration)),
	}
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &validClaims)
	validTokenString, _ := validToken.SignedString([]byte(config.Config.CacheServer.Key))
	validAuthHeader := "Bearer " + validTokenString

	invalidSignatureToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &validClaims)
	invalidSignatureTokenString, _ := invalidSignatureToken.SignedString([]byte("wrong-key"))
	invalidSignatureAuthHeader := "Bearer " + invalidSignatureTokenString

	expiredClaims := auth.Claims{
		ExpiresAt: jwt.NewNumericDate(now.Add(-time.Minute)),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte(config.Config.CacheServer.Key))
	expiredAuthHeader := "Bearer " + expiredTokenString

	revokedClaims := auth.Claims{
		IsRevoked: true,
		ExpiresAt: jwt.NewNumericDate(now.Add(-time.Minute)),
	}
	revokedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &revokedClaims)
	revokedTokenString, _ := revokedToken.SignedString([]byte(config.Config.CacheServer.Key))
	revokedTokenHeader := "Bearer " + revokedTokenString

	tests := []struct {
		name       string
		authHeader string
		wantClaims *auth.Claims
		wantErr    error
	}{
		{
			name:       "Valid token",
			authHeader: validAuthHeader,
			wantClaims: &validClaims,
			wantErr:    nil,
		},
		{
			name:       "Invalid token format - missing Bearer",
			authHeader: validTokenString,
			wantClaims: nil,
			wantErr:    auth.ErrInvalidTokenFormat,
		},
		{
			name:       "Invalid token format - too short",
			authHeader: "Bearer",
			wantClaims: nil,
			wantErr:    auth.ErrInvalidTokenFormat,
		},
		{
			name:       "Invalid signature",
			authHeader: invalidSignatureAuthHeader,
			wantClaims: nil,
			wantErr:    jwt.ErrSignatureInvalid,
		},
		{
			name:       "Expired token",
			authHeader: expiredAuthHeader,
			wantClaims: nil,
			wantErr:    jwt.ErrTokenExpired,
		},
		{
			name:       "Malformed token",
			authHeader: "Bearer malformed.token.string",
			wantClaims: nil,
			wantErr:    jwt.ErrTokenMalformed,
		},
		{
			name:       "Revoked token",
			authHeader: revokedTokenHeader,
			wantClaims: nil,
			wantErr:    auth.ErrTokenRevoked,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClaims, gotErr := auth.ParseToken(tt.authHeader)

			if gotErr != nil {
				if tt.wantErr == nil || !errors.Is(gotErr, tt.wantErr) {
					t.Errorf("ParseToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr != nil {
				t.Errorf("ParseToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if tt.wantClaims != nil {
				if gotClaims == nil {
					t.Errorf("ParseToken() gotClaims = nil, want non-nil")
					return
				}
				if !reflect.DeepEqual(gotClaims, tt.wantClaims) {
					t.Errorf("ParseToken() gotClaims = %v, wantClaims %v", gotClaims, tt.wantClaims)
				}
			} else if gotClaims != nil {
				t.Errorf("ParseToken() gotClaims = %v, want nil", gotClaims)
			}
		})
	}
}

func TestGenerateTokens(t *testing.T) {
	config.LoadConfig()

	tests := []struct {
		name                     string
		duration                 time.Duration
		wantAccessTokenNonEmpty  bool
		wantRefreshTokenNonEmpty bool
		wantErr                  bool
	}{
		{
			name:                     "Valid duration",
			duration:                 1 * time.Minute,
			wantAccessTokenNonEmpty:  true,
			wantRefreshTokenNonEmpty: true,
			wantErr:                  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAccessToken, err := auth.GenerateJwtWithDuration("", tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantAccessTokenNonEmpty && gotAccessToken == "" {
				t.Errorf("GenerateTokens() gotAccessToken = %q, want non-empty", gotAccessToken)
			}

			if gotAccessToken != "" {
				token, _, err := new(jwt.Parser).ParseUnverified(gotAccessToken, &auth.Claims{})
				if err != nil {
					t.Errorf("GenerateTokens() generated invalid access token: %+v", err)
					return
				}
				if claims, ok := token.Claims.(*auth.Claims); ok {
					if !claims.ExpiresAt.After(time.Now().Add(tt.duration-time.Minute)) || !claims.ExpiresAt.Before(time.Now().Add(tt.duration+time.Hour)) {
						t.Errorf("GenerateTokens() access token expiry is not within expected range: %+v", time.Until(claims.ExpiresAt.Time))
					}
				} else {
					t.Errorf("GenerateTokens() failed to parse access token claims")
				}
			}
		})
	}
}
