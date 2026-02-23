package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	ErrHeaderMissing = errors.New("error Authorization header is missing")
	ErrWrongToken    = errors.New("error token mismatch")
)

// Protect protects routes allowing access
// only checks for the validity of tokens
func Protect(token string) gin.HandlerFunc {
	tokenC := new(Claims)
	err := parseToken(tokenC, token)
	if err != nil {
		zap.S().Panicf("failed to protect endpoint bad token: %s, err: %v", token, err)
		return nil
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.String(http.StatusUnauthorized, ErrHeaderMissing.Error())
			c.Abort()
			return
		}

		claims, err := ParseToken(authHeader)
		if err != nil {
			zap.S().Errorf("Auth failed with err = %+v", err)
			c.String(http.StatusUnauthorized, "%v", errors.Join(ErrTokenNotValid, err))
			c.Abort()
			return
		}

		if tokenC.Id != claims.Id {
			zap.S().Errorf("Error token ids don't match", ErrWrongToken)
			c.String(http.StatusUnauthorized, "%v", ErrWrongToken)
			c.Abort()
			return
		}

		c.Next()
	}
}
