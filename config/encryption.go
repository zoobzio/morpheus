package config

import (
	"encoding/hex"
	"fmt"

	"github.com/zoobzio/check"
)

// Encryption holds configuration for field-level encryption.
type Encryption struct {
	// AESKey is a 64-character hex-encoded AES-256 key (32 bytes).
	AESKey string `env:"MORPHEUS_ENCRYPTION_KEY"`
}

// Validate validates the Encryption configuration.
func (c Encryption) Validate() error {
	if err := check.All(
		check.Str(c.AESKey, "aes_key").Required().MinLen(64).MaxLen(64).V(),
	).Err(); err != nil {
		return err
	}
	if _, err := hex.DecodeString(c.AESKey); err != nil {
		return fmt.Errorf("aes_key: must be a valid hex-encoded string: %w", err)
	}
	return nil
}

// Key decodes the hex-encoded AES key and returns the raw 32-byte key.
func (c Encryption) Key() []byte {
	b, _ := hex.DecodeString(c.AESKey)
	return b
}
