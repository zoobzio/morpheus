package models

import (
	"testing"
	"time"
)

func TestSession_Validate_Success(t *testing.T) {
	s := Session{
		Token:  "test_token_abc123",
		UserID: "01942d3a-1234-7abc-8def-0123456789ab",
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestSession_Validate_MissingToken(t *testing.T) {
	s := Session{
		UserID: "01942d3a-1234-7abc-8def-0123456789ab",
	}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for missing Token, got nil")
	}
}

func TestSession_Validate_MissingUserID(t *testing.T) {
	s := Session{
		Token: "test_token_abc123",
	}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for missing UserID, got nil")
	}
}

func TestSession_IsExpired_NotExpired(t *testing.T) {
	s := Session{
		Token:     "test_token_abc123",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if s.IsExpired() {
		t.Error("expected session to be active, IsExpired returned true")
	}
}

func TestSession_IsExpired_Expired(t *testing.T) {
	s := Session{
		Token:     "test_token_abc123",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		ExpiresAt: time.Now().Add(-time.Second),
	}
	if !s.IsExpired() {
		t.Error("expected session to be expired, IsExpired returned false")
	}
}

func TestSession_IsExpired_ExpiresNow(t *testing.T) {
	// A session expiring exactly now is considered expired when
	// time.Now().After(ExpiresAt) is evaluated — this is an
	// edge case that may vary depending on clock resolution.
	// We test a moment clearly in the past to ensure determinism.
	s := Session{
		Token:     "test_token_abc123",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		ExpiresAt: time.Now().Add(-time.Millisecond),
	}
	if !s.IsExpired() {
		t.Error("expected session 1ms past expiry to be expired")
	}
}

func TestSession_Clone(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	s := Session{
		Token:     "test_token_abc123",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		CreatedAt: now,
		ExpiresAt: now.Add(168 * time.Hour),
	}
	c := s.Clone()

	if c.Token != s.Token {
		t.Errorf("Token mismatch: got %q want %q", c.Token, s.Token)
	}
	if c.UserID != s.UserID {
		t.Errorf("UserID mismatch: got %q want %q", c.UserID, s.UserID)
	}
	if !c.CreatedAt.Equal(s.CreatedAt) {
		t.Errorf("CreatedAt mismatch: got %v want %v", c.CreatedAt, s.CreatedAt)
	}
	if !c.ExpiresAt.Equal(s.ExpiresAt) {
		t.Errorf("ExpiresAt mismatch: got %v want %v", c.ExpiresAt, s.ExpiresAt)
	}
}

func TestSession_Clone_Independence(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	s := Session{
		Token:     "test_token_abc123",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		CreatedAt: now,
		ExpiresAt: now.Add(168 * time.Hour),
	}
	c := s.Clone()

	// Mutate the clone; original must be unaffected.
	c.Token = "mutated_token"

	if s.Token != "test_token_abc123" {
		t.Errorf("original Token was mutated: got %q", s.Token)
	}
}
