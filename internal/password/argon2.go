// Package password provides Argon2id password hashing and verification.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Parameters holds the Argon2id configuration.
type Parameters struct {
	Memory      uint32
	Time        uint32
	Parallelism uint8
	SaltLen     uint32
	KeyLen      uint32
}

// defaults are the secure parameter defaults for Argon2id.
var defaults = Parameters{
	Memory:      65536,
	Time:        3,
	Parallelism: 4,
	SaltLen:     16,
	KeyLen:      32,
}

// ErrMalformedHash is returned by Verify when the encoded hash string is invalid.
var ErrMalformedHash = errors.New("password: malformed hash")

// Hash hashes password using Argon2id with secure defaults.
// The returned string is a self-contained encoded hash that includes the salt
// and parameters, formatted as:
//
//	$argon2id$v=19$m=65536,t=3,p=4$<base64-salt>$<base64-hash>
func Hash(password string) (string, error) {
	salt := make([]byte, defaults.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: generating salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		defaults.Time,
		defaults.Memory,
		defaults.Parallelism,
		defaults.KeyLen,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		defaults.Memory,
		defaults.Time,
		defaults.Parallelism,
		encodedSalt,
		encodedHash,
	)

	return encoded, nil
}

// Verify compares password against an encoded hash string produced by Hash.
// Returns true if the password matches, false if it does not.
// Returns ErrMalformedHash if the encoded string cannot be parsed.
func Verify(password, hash string) (bool, error) {
	p, salt, hashBytes, err := parse(hash)
	if err != nil {
		return false, err
	}

	candidate := argon2.IDKey(
		[]byte(password),
		salt,
		p.Time,
		p.Memory,
		p.Parallelism,
		p.KeyLen,
	)

	if subtle.ConstantTimeCompare(hashBytes, candidate) == 1 {
		return true, nil
	}
	return false, nil
}

// parse decodes an encoded Argon2id hash string into its components.
func parse(encoded string) (Parameters, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// Expected: ["", "argon2id", "v=19", "m=65536,t=3,p=4", "<salt>", "<hash>"]
	if len(parts) != 6 {
		return Parameters{}, nil, nil, ErrMalformedHash
	}

	if parts[1] != "argon2id" {
		return Parameters{}, nil, nil, ErrMalformedHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return Parameters{}, nil, nil, ErrMalformedHash
	}
	if version != argon2.Version {
		return Parameters{}, nil, nil, ErrMalformedHash
	}

	var p Parameters
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Parallelism); err != nil {
		return Parameters{}, nil, nil, ErrMalformedHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Parameters{}, nil, nil, ErrMalformedHash
	}

	hashBytes, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Parameters{}, nil, nil, ErrMalformedHash
	}

	p.SaltLen = uint32(len(salt))
	p.KeyLen = uint32(len(hashBytes))

	return p, salt, hashBytes, nil
}
