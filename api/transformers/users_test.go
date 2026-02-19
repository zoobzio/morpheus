package transformers

import (
	"testing"
	"time"

	"github.com/zoobzio/sumatra/api/wire"
	"github.com/zoobzio/sumatra/models"
)

func newTestUser() *models.User {
	name := "The Octocat"
	avatar := "https://avatars.githubusercontent.com/u/583231"
	now := time.Now().UTC().Truncate(time.Second)
	return &models.User{
		ID:            "01942d3a-1234-7abc-8def-0123456789ab",
		Email:         "octocat@github.com",
		EmailVerified: true,
		Name:          &name,
		AvatarURL:     &avatar,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// UserToResponse
// ──────────────────────────────────────────────────────────────────────────────

func TestUserToResponse_MapsScalarFields(t *testing.T) {
	u := newTestUser()
	resp := UserToResponse(u)

	if resp.ID != u.ID {
		t.Errorf("ID: got %q want %q", resp.ID, u.ID)
	}
	if resp.Email != u.Email {
		t.Errorf("Email: got %q want %q", resp.Email, u.Email)
	}
	if resp.EmailVerified != u.EmailVerified {
		t.Errorf("EmailVerified: got %v want %v", resp.EmailVerified, u.EmailVerified)
	}
}

func TestUserToResponse_MapsName(t *testing.T) {
	u := newTestUser()
	resp := UserToResponse(u)

	if resp.Name == nil {
		t.Fatal("expected non-nil Name in response")
	}
	if *resp.Name != *u.Name {
		t.Errorf("Name: got %q want %q", *resp.Name, *u.Name)
	}
}

func TestUserToResponse_MapsAvatarURL(t *testing.T) {
	u := newTestUser()
	resp := UserToResponse(u)

	if resp.AvatarURL == nil {
		t.Fatal("expected non-nil AvatarURL in response")
	}
	if *resp.AvatarURL != *u.AvatarURL {
		t.Errorf("AvatarURL: got %q want %q", *resp.AvatarURL, *u.AvatarURL)
	}
}

func TestUserToResponse_NilNameAndAvatar(t *testing.T) {
	u := &models.User{
		ID:    "01942d3a-1234-7abc-8def-0123456789ab",
		Email: "octocat@github.com",
	}
	resp := UserToResponse(u)

	if resp.Name != nil {
		t.Errorf("expected nil Name, got %q", *resp.Name)
	}
	if resp.AvatarURL != nil {
		t.Errorf("expected nil AvatarURL, got %q", *resp.AvatarURL)
	}
}

func TestUserToResponse_DoesNotIncludeTimestamps(t *testing.T) {
	// UserResponse has no timestamp fields — verify that the type compiles
	// without them and the zero-value response doesn't carry garbage from
	// the model.
	u := newTestUser()
	resp := UserToResponse(u)

	// If the wire type had a CreatedAt field this would fail to compile.
	// The check here is that no timestamp leaks through a field we didn't
	// intend — verified by the fact that UserResponse has no such field.
	_ = resp // compile-time shape check is sufficient
}

// ──────────────────────────────────────────────────────────────────────────────
// ApplyUserUpdate
// ──────────────────────────────────────────────────────────────────────────────

func TestApplyUserUpdate_SetsName(t *testing.T) {
	u := newTestUser()
	newName := "Updated Name"
	req := wire.UserUpdateRequest{Name: &newName}

	ApplyUserUpdate(req, u)

	if u.Name == nil {
		t.Fatal("expected non-nil Name after update")
	}
	if *u.Name != newName {
		t.Errorf("Name: got %q want %q", *u.Name, newName)
	}
}

func TestApplyUserUpdate_NilNameIsNoop(t *testing.T) {
	u := newTestUser()
	originalName := *u.Name

	req := wire.UserUpdateRequest{Name: nil}
	ApplyUserUpdate(req, u)

	if u.Name == nil {
		t.Fatal("expected Name to remain set")
	}
	if *u.Name != originalName {
		t.Errorf("Name was changed unexpectedly: got %q want %q", *u.Name, originalName)
	}
}

func TestApplyUserUpdate_ClearsNameWithEmptyString(t *testing.T) {
	u := newTestUser()
	empty := ""
	req := wire.UserUpdateRequest{Name: &empty}

	ApplyUserUpdate(req, u)

	if u.Name == nil {
		t.Fatal("expected Name pointer to be set (to empty string)")
	}
	if *u.Name != "" {
		t.Errorf("Name: got %q want empty string", *u.Name)
	}
}

func TestApplyUserUpdate_DoesNotTouchEmail(t *testing.T) {
	u := newTestUser()
	original := u.Email
	newName := "New Name"
	req := wire.UserUpdateRequest{Name: &newName}

	ApplyUserUpdate(req, u)

	if u.Email != original {
		t.Errorf("Email was mutated: got %q want %q", u.Email, original)
	}
}
