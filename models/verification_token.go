package models

import (
	"time"

	"github.com/zoobzio/check"
)

// TokenType identifies the purpose of a verification token.
type TokenType string

const (
	// TokenTypeEmailVerify is issued to verify a new email address.
	TokenTypeEmailVerify TokenType = "email_verify"
	// TokenTypeMagicLink is issued for a passwordless sign-in link.
	TokenTypeMagicLink TokenType = "magic_link"
	// TokenTypePasswordReset is issued to authorise a password reset.
	TokenTypePasswordReset TokenType = "password_reset"
)

// VerificationToken is a short-lived, single-use token for email verification,
// magic-link sign-in, or password reset flows. Tokens are stored in Redis with
// a TTL derived from their type.
type VerificationToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Type      TokenType `json:"type"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired reports whether the token has passed its expiry time.
func (v VerificationToken) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

// Validate validates the VerificationToken model.
func (v VerificationToken) Validate() error {
	return check.All(
		check.Str(v.Token, "token").Required().V(),
		check.Str(v.UserID, "user_id").Required().V(),
		check.Str(string(v.Type), "type").Required().OneOf([]string{
			string(TokenTypeEmailVerify),
			string(TokenTypeMagicLink),
			string(TokenTypePasswordReset),
		}).V(),
	).Err()
}

// Clone returns a deep copy of the VerificationToken.
func (v VerificationToken) Clone() VerificationToken {
	return v
}
