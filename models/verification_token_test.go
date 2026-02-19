package models

import (
	"testing"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// Validate
// ──────────────────────────────────────────────────────────────────────────────

func TestVerificationToken_Validate_Success(t *testing.T) {
	v := VerificationToken{
		Token:  "abcdef1234567890",
		UserID: "01942d3a-1234-7abc-8def-0123456789ab",
		Type:   TokenTypeEmailVerify,
	}
	if err := v.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestVerificationToken_Validate_AllTypes(t *testing.T) {
	types := []TokenType{TokenTypeEmailVerify, TokenTypeMagicLink, TokenTypePasswordReset}
	for _, tt := range types {
		v := VerificationToken{
			Token:  "tok",
			UserID: "uid",
			Type:   tt,
		}
		if err := v.Validate(); err != nil {
			t.Errorf("expected no error for type %q, got: %v", tt, err)
		}
	}
}

func TestVerificationToken_Validate_MissingToken(t *testing.T) {
	v := VerificationToken{
		UserID: "uid",
		Type:   TokenTypeEmailVerify,
	}
	if err := v.Validate(); err == nil {
		t.Fatal("expected error for missing Token, got nil")
	}
}

func TestVerificationToken_Validate_MissingUserID(t *testing.T) {
	v := VerificationToken{
		Token: "tok",
		Type:  TokenTypeMagicLink,
	}
	if err := v.Validate(); err == nil {
		t.Fatal("expected error for missing UserID, got nil")
	}
}

func TestVerificationToken_Validate_MissingType(t *testing.T) {
	v := VerificationToken{
		Token:  "tok",
		UserID: "uid",
	}
	if err := v.Validate(); err == nil {
		t.Fatal("expected error for missing Type, got nil")
	}
}

func TestVerificationToken_Validate_InvalidType(t *testing.T) {
	v := VerificationToken{
		Token:  "tok",
		UserID: "uid",
		Type:   TokenType("unknown"),
	}
	if err := v.Validate(); err == nil {
		t.Fatal("expected error for invalid Type, got nil")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// IsExpired
// ──────────────────────────────────────────────────────────────────────────────

func TestVerificationToken_IsExpired_NotExpired(t *testing.T) {
	v := VerificationToken{
		Token:     "tok",
		UserID:    "uid",
		Type:      TokenTypeEmailVerify,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if v.IsExpired() {
		t.Error("expected token to be active, IsExpired returned true")
	}
}

func TestVerificationToken_IsExpired_Expired(t *testing.T) {
	v := VerificationToken{
		Token:     "tok",
		UserID:    "uid",
		Type:      TokenTypeMagicLink,
		ExpiresAt: time.Now().Add(-time.Second),
	}
	if !v.IsExpired() {
		t.Error("expected token to be expired, IsExpired returned false")
	}
}

func TestVerificationToken_IsExpired_JustExpired(t *testing.T) {
	v := VerificationToken{
		Token:     "tok",
		UserID:    "uid",
		Type:      TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(-time.Millisecond),
	}
	if !v.IsExpired() {
		t.Error("expected token 1ms past expiry to be expired")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Clone
// ──────────────────────────────────────────────────────────────────────────────

func TestVerificationToken_Clone_FieldEquality(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	v := VerificationToken{
		Token:     "abcdef1234567890",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		Type:      TokenTypeEmailVerify,
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}
	c := v.Clone()

	if c.Token != v.Token {
		t.Errorf("Token mismatch: got %q want %q", c.Token, v.Token)
	}
	if c.UserID != v.UserID {
		t.Errorf("UserID mismatch: got %q want %q", c.UserID, v.UserID)
	}
	if c.Type != v.Type {
		t.Errorf("Type mismatch: got %q want %q", c.Type, v.Type)
	}
	if !c.ExpiresAt.Equal(v.ExpiresAt) {
		t.Errorf("ExpiresAt mismatch: got %v want %v", c.ExpiresAt, v.ExpiresAt)
	}
	if !c.CreatedAt.Equal(v.CreatedAt) {
		t.Errorf("CreatedAt mismatch: got %v want %v", c.CreatedAt, v.CreatedAt)
	}
}

func TestVerificationToken_Clone_Independence(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	v := VerificationToken{
		Token:     "original_token",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		Type:      TokenTypeMagicLink,
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
	}
	c := v.Clone()

	c.Token = "mutated_token"
	c.Type = TokenTypePasswordReset

	if v.Token != "original_token" {
		t.Errorf("original Token was mutated: got %q", v.Token)
	}
	if v.Type != TokenTypeMagicLink {
		t.Errorf("original Type was mutated: got %q", v.Type)
	}
}
