package stores

import (
	"strings"
	"testing"
)

// ──────────────────────────────────────────────────────────────────────────────
// Key helper functions
// ──────────────────────────────────────────────────────────────────────────────

func TestSessionKey_Format(t *testing.T) {
	key := sessionKey("abc123")
	want := "session:abc123"
	if key != want {
		t.Errorf("sessionKey: got %q want %q", key, want)
	}
}

func TestSessionKey_EmptyToken(t *testing.T) {
	key := sessionKey("")
	want := "session:"
	if key != want {
		t.Errorf("sessionKey empty token: got %q want %q", key, want)
	}
}

func TestSessionKey_HasPrefix(t *testing.T) {
	key := sessionKey("tok")
	if !strings.HasPrefix(key, sessionPrefix) {
		t.Errorf("sessionKey should start with %q: got %q", sessionPrefix, key)
	}
}

func TestUserIndexKey_Format(t *testing.T) {
	key := userIndexKey("user-1", "tok-abc")
	want := "user_sessions:user-1:tok-abc"
	if key != want {
		t.Errorf("userIndexKey: got %q want %q", key, want)
	}
}

func TestUserIndexKey_HasPrefix(t *testing.T) {
	key := userIndexKey("uid", "tok")
	if !strings.HasPrefix(key, userIndexPrefix) {
		t.Errorf("userIndexKey should start with %q: got %q", userIndexPrefix, key)
	}
}

func TestUserIndexKey_ContainsUserID(t *testing.T) {
	userID := "01942d3a-1234-7abc-8def-0123456789ab"
	key := userIndexKey(userID, "tok")
	if !strings.Contains(key, userID) {
		t.Errorf("userIndexKey should contain userID %q: got %q", userID, key)
	}
}

func TestUserIndexKey_ContainsToken(t *testing.T) {
	token := "mytesttoken"
	key := userIndexKey("uid", token)
	if !strings.Contains(key, token) {
		t.Errorf("userIndexKey should contain token %q: got %q", token, key)
	}
}

func TestUserIndexKey_DifferentUsersProduceDifferentKeys(t *testing.T) {
	k1 := userIndexKey("user-1", "token")
	k2 := userIndexKey("user-2", "token")
	if k1 == k2 {
		t.Errorf("different users should produce different keys: both %q", k1)
	}
}

func TestUserIndexKey_DifferentTokensProduceDifferentKeys(t *testing.T) {
	k1 := userIndexKey("user-1", "token-a")
	k2 := userIndexKey("user-1", "token-b")
	if k1 == k2 {
		t.Errorf("different tokens should produce different keys: both %q", k1)
	}
}

func TestUserIndexScanPrefix_Format(t *testing.T) {
	prefix := userIndexScanPrefix("user-1")
	want := "user_sessions:user-1:"
	if prefix != want {
		t.Errorf("userIndexScanPrefix: got %q want %q", prefix, want)
	}
}

func TestUserIndexScanPrefix_HasUserIndexPrefix(t *testing.T) {
	prefix := userIndexScanPrefix("uid")
	if !strings.HasPrefix(prefix, userIndexPrefix) {
		t.Errorf("scan prefix should start with %q: got %q", userIndexPrefix, prefix)
	}
}

func TestUserIndexScanPrefix_EndsWithColon(t *testing.T) {
	prefix := userIndexScanPrefix("uid")
	if !strings.HasSuffix(prefix, ":") {
		t.Errorf("scan prefix should end with ':': got %q", prefix)
	}
}

func TestUserIndexScanPrefix_MatchesUserIndexKeyPrefix(t *testing.T) {
	// Keys produced by userIndexKey must start with the scan prefix for the
	// same userID — this is the invariant the ListByUser/DeleteByUser logic
	// relies on.
	userID := "user-abc"
	token := "tok-xyz"
	key := userIndexKey(userID, token)
	scanPrefix := userIndexScanPrefix(userID)

	if !strings.HasPrefix(key, scanPrefix) {
		t.Errorf("userIndexKey %q should start with scanPrefix %q", key, scanPrefix)
	}
}

func TestUserIndexScanPrefix_DoesNotMatchOtherUser(t *testing.T) {
	key := userIndexKey("user-1", "tok")
	scanPrefix := userIndexScanPrefix("user-2")

	if strings.HasPrefix(key, scanPrefix) {
		t.Errorf("user-1 key %q should NOT match user-2 scan prefix %q", key, scanPrefix)
	}
}
