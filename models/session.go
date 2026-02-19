package models

import (
	"time"

	"github.com/zoobzio/check"
)

// Session represents an authenticated user session stored in Redis.
type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsExpired reports whether the session has expired.
func (s Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Validate validates the Session model.
func (s Session) Validate() error {
	return check.All(
		check.Str(s.Token, "token").Required().V(),
		check.Str(s.UserID, "user_id").Required().V(),
	).Err()
}

// Clone returns a deep copy of the Session.
func (s Session) Clone() Session {
	return s
}
