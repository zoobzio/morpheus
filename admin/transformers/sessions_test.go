package transformers

import (
	"strings"
	"testing"
	"time"

	"github.com/zoobzio/sumatra/models"
)

func newTestSession() *models.Session {
	now := time.Now().UTC().Truncate(time.Second)
	return &models.Session{
		Token:     "abcdefghijklmnopqrstuvwxyz012345",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		CreatedAt: now,
		ExpiresAt: now.Add(168 * time.Hour),
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// maskToken (internal helper — tested via SessionToAdminResponse)
// ──────────────────────────────────────────────────────────────────────────────

func TestMaskToken_LongToken(t *testing.T) {
	tok := "abcdefghijklmnopqrstuvwxyz"
	masked := maskToken(tok)

	if !strings.HasSuffix(masked, "...") {
		t.Errorf("masked token should end with '...': got %q", masked)
	}
	if !strings.HasPrefix(masked, tok[:8]) {
		t.Errorf("masked token should start with first 8 chars: got %q", masked)
	}
	if len(masked) != 11 { // 8 chars + "..."
		t.Errorf("masked token length: got %d want 11", len(masked))
	}
}

func TestMaskToken_ExactlyEightChars(t *testing.T) {
	tok := "12345678"
	masked := maskToken(tok)

	// len == 8, so the code takes the else branch: token[:8] + "..."
	if masked != "12345678..." {
		t.Errorf("got %q want %q", masked, "12345678...")
	}
}

func TestMaskToken_ShortToken(t *testing.T) {
	tok := "abc"
	masked := maskToken(tok)

	// len <= 8, so entire token + "..."
	if masked != "abc..." {
		t.Errorf("got %q want %q", masked, "abc...")
	}
}

func TestMaskToken_EmptyToken(t *testing.T) {
	masked := maskToken("")

	if masked != "..." {
		t.Errorf("got %q want %q", masked, "...")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// SessionToAdminResponse
// ──────────────────────────────────────────────────────────────────────────────

func TestSessionToAdminResponse_TokenIsMasked(t *testing.T) {
	s := newTestSession()
	resp := SessionToAdminResponse(s)

	// Token must not be the original.
	if resp.Token == s.Token {
		t.Error("expected token to be masked, but got the full token")
	}
	if !strings.HasSuffix(resp.Token, "...") {
		t.Errorf("masked token should end with '...': got %q", resp.Token)
	}
}

func TestSessionToAdminResponse_TokenPrefix(t *testing.T) {
	s := newTestSession()
	resp := SessionToAdminResponse(s)

	if !strings.HasPrefix(resp.Token, s.Token[:8]) {
		t.Errorf("masked token should start with first 8 chars of original: got %q", resp.Token)
	}
}

func TestSessionToAdminResponse_MapsUserID(t *testing.T) {
	s := newTestSession()
	resp := SessionToAdminResponse(s)

	if resp.UserID != s.UserID {
		t.Errorf("UserID: got %q want %q", resp.UserID, s.UserID)
	}
}

func TestSessionToAdminResponse_MapsTimestamps(t *testing.T) {
	s := newTestSession()
	resp := SessionToAdminResponse(s)

	if !resp.CreatedAt.Equal(s.CreatedAt) {
		t.Errorf("CreatedAt: got %v want %v", resp.CreatedAt, s.CreatedAt)
	}
	if !resp.ExpiresAt.Equal(s.ExpiresAt) {
		t.Errorf("ExpiresAt: got %v want %v", resp.ExpiresAt, s.ExpiresAt)
	}
}

func TestSessionToAdminResponse_ShortTokenMasked(t *testing.T) {
	s := &models.Session{
		Token:     "short",
		UserID:    "01942d3a-1234-7abc-8def-0123456789ab",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	resp := SessionToAdminResponse(s)

	if resp.Token != "short..." {
		t.Errorf("short token: got %q want %q", resp.Token, "short...")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// SessionsToAdminList
// ──────────────────────────────────────────────────────────────────────────────

func TestSessionsToAdminList_EmptySlice(t *testing.T) {
	resp := SessionsToAdminList([]*models.Session{})

	if len(resp.Sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(resp.Sessions))
	}
}

func TestSessionsToAdminList_NilSlice(t *testing.T) {
	resp := SessionsToAdminList(nil)

	if len(resp.Sessions) != 0 {
		t.Errorf("expected 0 sessions for nil input, got %d", len(resp.Sessions))
	}
}

func TestSessionsToAdminList_MasksAllTokens(t *testing.T) {
	sessions := []*models.Session{
		newTestSession(),
		{Token: "anotherlongtoken123", UserID: "user-2", CreatedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)},
	}
	resp := SessionsToAdminList(sessions)

	if len(resp.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(resp.Sessions))
	}
	for i, s := range sessions {
		if resp.Sessions[i].Token == s.Token {
			t.Errorf("sessions[%d].Token was not masked", i)
		}
		if !strings.HasSuffix(resp.Sessions[i].Token, "...") {
			t.Errorf("sessions[%d].Token should end with '...': got %q", i, resp.Sessions[i].Token)
		}
	}
}

func TestSessionsToAdminList_MapsUserIDs(t *testing.T) {
	s1 := newTestSession()
	s1.UserID = "user-id-one"
	s2 := newTestSession()
	s2.UserID = "user-id-two"

	resp := SessionsToAdminList([]*models.Session{s1, s2})

	if resp.Sessions[0].UserID != "user-id-one" {
		t.Errorf("sessions[0].UserID: got %q want %q", resp.Sessions[0].UserID, "user-id-one")
	}
	if resp.Sessions[1].UserID != "user-id-two" {
		t.Errorf("sessions[1].UserID: got %q want %q", resp.Sessions[1].UserID, "user-id-two")
	}
}
