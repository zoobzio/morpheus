package models

import (
	"testing"
	"time"
)

func TestUser_Validate_Success(t *testing.T) {
	u := User{
		ID:    "01942d3a-1234-7abc-8def-0123456789ab",
		Email: "octocat@github.com",
	}
	if err := u.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestUser_Validate_MissingID(t *testing.T) {
	u := User{
		Email: "octocat@github.com",
	}
	if err := u.Validate(); err == nil {
		t.Fatal("expected error for missing ID, got nil")
	}
}

func TestUser_Validate_MissingEmail(t *testing.T) {
	u := User{
		ID: "01942d3a-1234-7abc-8def-0123456789ab",
	}
	if err := u.Validate(); err == nil {
		t.Fatal("expected error for missing email, got nil")
	}
}

func TestUser_Clone_NilPointers(t *testing.T) {
	u := User{
		ID:        "01942d3a-1234-7abc-8def-0123456789ab",
		Email:     "octocat@github.com",
		Name:      nil,
		AvatarURL: nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	c := u.Clone()
	if c.Name != nil {
		t.Error("expected nil Name in clone")
	}
	if c.AvatarURL != nil {
		t.Error("expected nil AvatarURL in clone")
	}
	if c.PasswordHash != nil {
		t.Error("expected nil PasswordHash in clone")
	}
}

func TestUser_Clone_DeepCopiesPointers(t *testing.T) {
	name := "The Octocat"
	avatar := "https://avatars.githubusercontent.com/u/583231"
	hash := "$argon2id$v=19$m=65536,t=1,p=4$abc$xyz"
	u := User{
		ID:           "01942d3a-1234-7abc-8def-0123456789ab",
		Email:        "octocat@github.com",
		PasswordHash: &hash,
		Name:         &name,
		AvatarURL:    &avatar,
	}
	c := u.Clone()

	// Mutate the original pointer.
	newName := "Changed"
	u.Name = &newName

	if *c.Name != "The Octocat" {
		t.Errorf("clone Name was mutated by original change: got %q", *c.Name)
	}
}

func TestUser_Clone_PasswordHashDeepCopy(t *testing.T) {
	hash := "$argon2id$v=19$m=65536,t=1,p=4$abc$xyz"
	u := User{
		ID:           "01942d3a-1234-7abc-8def-0123456789ab",
		Email:        "octocat@github.com",
		PasswordHash: &hash,
	}
	c := u.Clone()

	// Mutate the clone's hash value.
	*c.PasswordHash = "changed"

	if *u.PasswordHash != hash {
		t.Errorf("original PasswordHash was mutated by clone change: got %q", *u.PasswordHash)
	}
}

func TestUser_Clone_AvatarURLDeepCopy(t *testing.T) {
	name := "The Octocat"
	avatar := "https://avatars.githubusercontent.com/u/583231"
	u := User{
		ID:        "01942d3a-1234-7abc-8def-0123456789ab",
		Email:     "octocat@github.com",
		Name:      &name,
		AvatarURL: &avatar,
	}
	c := u.Clone()

	// Mutate the cloned pointer value.
	*c.AvatarURL = "https://changed.example.com"

	if *u.AvatarURL != "https://avatars.githubusercontent.com/u/583231" {
		t.Errorf("original AvatarURL was mutated by clone change: got %q", *u.AvatarURL)
	}
}

func TestUser_Clone_ScalarFields(t *testing.T) {
	now := time.Now().UTC()
	u := User{
		ID:            "01942d3a-1234-7abc-8def-0123456789ab",
		Email:         "octocat@github.com",
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	c := u.Clone()

	if c.ID != u.ID {
		t.Errorf("ID mismatch: got %q want %q", c.ID, u.ID)
	}
	if c.Email != u.Email {
		t.Errorf("Email mismatch: got %q want %q", c.Email, u.Email)
	}
	if c.EmailVerified != u.EmailVerified {
		t.Errorf("EmailVerified mismatch: got %v want %v", c.EmailVerified, u.EmailVerified)
	}
	if !c.CreatedAt.Equal(u.CreatedAt) {
		t.Errorf("CreatedAt mismatch: got %v want %v", c.CreatedAt, u.CreatedAt)
	}
}
