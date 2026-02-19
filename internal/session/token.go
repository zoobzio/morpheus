// Package session provides session token generation and state cookie management.
package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateToken generates a cryptographically random 32-byte token encoded as base64url.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
