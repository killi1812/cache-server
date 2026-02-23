package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var ErrHeaderMissing = errors.New("err Authorization header is missing")

// Protect protects routes allowing access
// only checks for the validity of tokens
func Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.String(http.StatusUnauthorized, ErrHeaderMissing.Error())
			c.Abort()
			return
		}

		_, err := ParseToken(authHeader)
		if err != nil {
			zap.S().Errorf("Auth failed with err = %+v", err)
			c.String(http.StatusUnauthorized, "%v", errors.Join(ErrTokenNotValid, err))
			c.Abort()
			return
		}

		c.Next()
	}
}
