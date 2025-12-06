package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateAPIKey returns a cryptographically secure random string.
func GenerateAPIKey(bytesLen int) (string, error) {
	if bytesLen <= 0 {
		bytesLen = 32
	}
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate api key: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
