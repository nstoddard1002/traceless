package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateID creates a random 32-character hex string.
func GenerateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
