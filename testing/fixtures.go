//go:build testing

// Package testing provides test infrastructure: fixtures, mocks, and helpers.
package testing

import (
	"testing"
	"time"

	"github.com/zoobzio/sumatra/models"
)

// NewUser returns a User with sensible defaults for testing.
func NewUser(t *testing.T) *models.User {
	t.Helper()
	name := "The Octocat"
	avatar := "https://avatars.githubusercontent.com/u/583231"
	now := time.Now().UTC().Truncate(time.Second)
	return &models.User{
		ID:            "01942d3a-1234-7abc-8def-0123456789ab",
		Email:         "octocat@github.com",
		EmailVerified: false,
		Name:          &name,
		AvatarURL:     &avatar,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// NewUsers returns n Users with sensible defaults for testing.
// Each user gets a distinct ID and email to avoid uniqueness conflicts.
func NewUsers(t *testing.T, n int) []*models.User {
	t.Helper()
	users := make([]*models.User, n)
	for i := range users {
		u := NewUser(t)
		u.ID = "01942d3a-1234-7abc-8def-" + padInt(i)
		u.Email = "octocat" + padInt(i) + "@github.com"
		users[i] = u
	}
	return users
}

// NewProvider returns a Provider with sensible defaults for testing.
func NewProvider(t *testing.T) *models.Provider {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	return &models.Provider{
		ID:             1,
		UserID:         "01942d3a-1234-7abc-8def-0123456789ab",
		Type:           models.ProviderTypeGitHub,
		ProviderUserID: "583231",
		AccessToken:    "gho_test_access_token_value",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// NewSession returns a Session with sensible defaults for testing.
// The session is not expired by default.
func NewSession(t *testing.T) *models.Session {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	return &models.Session{
		Token:     "test_session_token_abcdefgh1234",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		CreatedAt: now,
		ExpiresAt: now.Add(168 * time.Hour),
	}
}

// NewExpiredSession returns a Session that has already expired.
func NewExpiredSession(t *testing.T) *models.Session {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	return &models.Session{
		Token:     "expired_session_token_abcdefgh",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		CreatedAt: now.Add(-48 * time.Hour),
		ExpiresAt: now.Add(-1 * time.Second),
	}
}

// padInt returns a zero-padded 12-digit decimal string for use in IDs.
func padInt(i int) string {
	const digits = "0123456789"
	buf := [12]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}
	pos := 11
	if i == 0 {
		return string(buf[:])
	}
	for i > 0 && pos >= 0 {
		buf[pos] = digits[i%10]
		i /= 10
		pos--
	}
	return string(buf[:])
}
