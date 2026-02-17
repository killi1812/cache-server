package auth

import (
	"crypto/rand"
	"encoding/hex"

	"go.uber.org/zap"
)

// GenerateToken creates a hex-encoded random string of length n
func GenerateTokenN(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	zap.S().Debugf("Generated a %d token", len(bytes))
	return hex.EncodeToString(bytes), nil
}

func GenerateToken() (string, error) {
	zap.S().Debugf("Generating a 32 bit token")
	return GenerateTokenN(32)
}
