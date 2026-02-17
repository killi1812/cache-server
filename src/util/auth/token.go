package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken creates a hex-encoded random string of length n
func GenerateTokenN(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateToken() (string, error) {
	return GenerateTokenN(32)
}
