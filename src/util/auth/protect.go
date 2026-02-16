package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Protect protects routes allowing access
// only checks for the validity of tokens
func Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Missing token")
			return
		}

		// TODO: check if claims needed for something
		token, _, err := ParseToken(authHeader)
		if err != nil {
			zap.S().Infof("Auth failed with err = %+v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token format")
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token")
			return
		}

		c.Next()
	}
}
