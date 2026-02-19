package transformers

import (
	"testing"
	"time"

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
// UserToAdminResponse
// ──────────────────────────────────────────────────────────────────────────────

func TestUserToAdminResponse_MapsScalarFields(t *testing.T) {
	u := newTestUser()
	resp := UserToAdminResponse(u)

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

func TestUserToAdminResponse_MapsTimestamps(t *testing.T) {
	u := newTestUser()
	resp := UserToAdminResponse(u)

	if !resp.CreatedAt.Equal(u.CreatedAt) {
		t.Errorf("CreatedAt: got %v want %v", resp.CreatedAt, u.CreatedAt)
	}
	if !resp.UpdatedAt.Equal(u.UpdatedAt) {
		t.Errorf("UpdatedAt: got %v want %v", resp.UpdatedAt, u.UpdatedAt)
	}
}

func TestUserToAdminResponse_MapsName(t *testing.T) {
	u := newTestUser()
	resp := UserToAdminResponse(u)

	if resp.Name == nil {
		t.Fatal("expected non-nil Name in response")
	}
	if *resp.Name != *u.Name {
		t.Errorf("Name: got %q want %q", *resp.Name, *u.Name)
	}
}

func TestUserToAdminResponse_MapsAvatarURL(t *testing.T) {
	u := newTestUser()
	resp := UserToAdminResponse(u)

	if resp.AvatarURL == nil {
		t.Fatal("expected non-nil AvatarURL in response")
	}
	if *resp.AvatarURL != *u.AvatarURL {
		t.Errorf("AvatarURL: got %q want %q", *resp.AvatarURL, *u.AvatarURL)
	}
}

func TestUserToAdminResponse_NilOptionals(t *testing.T) {
	u := &models.User{
		ID:        "01942d3a-1234-7abc-8def-0123456789ab",
		Email:     "octocat@github.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	resp := UserToAdminResponse(u)

	if resp.Name != nil {
		t.Errorf("expected nil Name, got %q", *resp.Name)
	}
	if resp.AvatarURL != nil {
		t.Errorf("expected nil AvatarURL, got %q", *resp.AvatarURL)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// UsersToAdminList
// ──────────────────────────────────────────────────────────────────────────────

func TestUsersToAdminList_EmptySlice(t *testing.T) {
	resp := UsersToAdminList([]*models.User{}, 0)

	if len(resp.Users) != 0 {
		t.Errorf("expected 0 users, got %d", len(resp.Users))
	}
	if resp.Total != 0 {
		t.Errorf("Total: got %d want 0", resp.Total)
	}
}

func TestUsersToAdminList_Total(t *testing.T) {
	u := newTestUser()
	resp := UsersToAdminList([]*models.User{u}, 42)

	if resp.Total != 42 {
		t.Errorf("Total: got %d want 42", resp.Total)
	}
}

func TestUsersToAdminList_MapsAllUsers(t *testing.T) {
	users := make([]*models.User, 3)
	for i := range users {
		users[i] = newTestUser()
		users[i].Email = "user" + string(rune('A'+i)) + "@example.com"
	}
	resp := UsersToAdminList(users, len(users))

	if len(resp.Users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(resp.Users))
	}
	for i, u := range users {
		if resp.Users[i].Email != u.Email {
			t.Errorf("users[%d].Email: got %q want %q", i, resp.Users[i].Email, u.Email)
		}
	}
}

func TestUsersToAdminList_TotalCanDifferFromSliceLen(t *testing.T) {
	// Total represents the DB count, not the page length.
	u := newTestUser()
	resp := UsersToAdminList([]*models.User{u}, 1000)

	if len(resp.Users) != 1 {
		t.Errorf("expected 1 user in page, got %d", len(resp.Users))
	}
	if resp.Total != 1000 {
		t.Errorf("Total: got %d want 1000", resp.Total)
	}
}

func TestUsersToAdminList_NilSlice(t *testing.T) {
	resp := UsersToAdminList(nil, 0)

	if len(resp.Users) != 0 {
		t.Errorf("expected 0 users for nil input, got %d", len(resp.Users))
	}
}
