package session

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestManager(t *testing.T) *StateManager {
	t.Helper()
	// Secret must be at least 32 chars to satisfy production validation.
	return NewStateManager("a-test-secret-that-is-long-enough!!", "localhost", false)
}

// requestWithCookie returns an *http.Request carrying the named cookie.
func requestWithCookie(name, value string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/callback", nil)
	r.AddCookie(&http.Cookie{Name: name, Value: value})
	return r
}

func TestGenerateState_ReturnsStateAndCookie(t *testing.T) {
	m := newTestManager(t)

	state, cookie, err := m.GenerateState()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if state == "" {
		t.Error("expected non-empty state string")
	}
	if cookie == nil {
		t.Fatal("expected non-nil cookie")
	}
	if cookie.Name != stateCookieName {
		t.Errorf("cookie name: got %q want %q", cookie.Name, stateCookieName)
	}
	if cookie.Value != state {
		t.Errorf("cookie value does not match state: got %q want %q", cookie.Value, state)
	}
}

func TestGenerateState_CookieAttributes(t *testing.T) {
	m := NewStateManager("a-test-secret-that-is-long-enough!!", "example.com", true)

	_, cookie, err := m.GenerateState()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !cookie.Secure {
		t.Error("expected Secure=true")
	}
	if !cookie.HttpOnly {
		t.Error("expected HttpOnly=true")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite: got %v want SameSiteLaxMode", cookie.SameSite)
	}
	if cookie.Domain != "example.com" {
		t.Errorf("Domain: got %q want %q", cookie.Domain, "example.com")
	}
	if cookie.MaxAge != int(stateTTL.Seconds()) {
		t.Errorf("MaxAge: got %d want %d", cookie.MaxAge, int(stateTTL.Seconds()))
	}
	if cookie.Path != "/" {
		t.Errorf("Path: got %q want %q", cookie.Path, "/")
	}
}

func TestGenerateState_HasThreeParts(t *testing.T) {
	// State format: nonce|expiry.signature
	m := newTestManager(t)
	state, _, err := m.GenerateState()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	lastDot := strings.LastIndex(state, ".")
	if lastDot < 0 {
		t.Fatalf("state has no '.': %q", state)
	}
	payload := state[:lastDot]
	if !strings.Contains(payload, "|") {
		t.Errorf("payload has no '|': %q", payload)
	}
}

func TestValidateState_RoundTrip(t *testing.T) {
	m := newTestManager(t)

	state, _, err := m.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState: %v", err)
	}

	r := requestWithCookie(stateCookieName, state)
	if err := m.ValidateState(state, r); err != nil {
		t.Fatalf("expected valid state, got: %v", err)
	}
}

func TestValidateState_NoCookie(t *testing.T) {
	m := newTestManager(t)

	state, _, err := m.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/callback", nil) // no cookie
	if err := m.ValidateState(state, r); err != ErrInvalidState {
		t.Errorf("expected ErrInvalidState, got: %v", err)
	}
}

func TestValidateState_CookieMismatch(t *testing.T) {
	m := newTestManager(t)

	state, _, err := m.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState: %v", err)
	}

	other, _, err := m.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState (other): %v", err)
	}

	// Cookie holds `state`, but param is `other` — mismatch.
	r := requestWithCookie(stateCookieName, state)
	if err := m.ValidateState(other, r); err != ErrInvalidState {
		t.Errorf("expected ErrInvalidState on mismatch, got: %v", err)
	}
}

func TestValidateState_TamperedSignature(t *testing.T) {
	m := newTestManager(t)

	state, _, err := m.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState: %v", err)
	}

	// Flip the last character of the signature.
	tampered := state[:len(state)-1] + "X"
	if tampered[len(tampered)-1] == state[len(state)-1] {
		tampered = state[:len(state)-1] + "Y"
	}

	r := requestWithCookie(stateCookieName, tampered)
	if err := m.ValidateState(tampered, r); err != ErrInvalidState {
		t.Errorf("expected ErrInvalidState for tampered signature, got: %v", err)
	}
}

func TestValidateState_TamperedPayload(t *testing.T) {
	m := newTestManager(t)

	state, _, err := m.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState: %v", err)
	}

	// Replace the nonce portion of the payload while keeping the original sig.
	lastDot := strings.LastIndex(state, ".")
	sig := state[lastDot:]
	payload := "fakefakenonce|" + strings.Split(state[:lastDot], "|")[1]
	tampered := payload + sig

	r := requestWithCookie(stateCookieName, tampered)
	if err := m.ValidateState(tampered, r); err != ErrInvalidState {
		t.Errorf("expected ErrInvalidState for tampered payload, got: %v", err)
	}
}

func TestValidateState_Expired(t *testing.T) {
	m := newTestManager(t)

	state, _, err := m.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState: %v", err)
	}

	// Reconstruct the state with an expiry in the past.
	// We re-sign it with the same manager to isolate the expiry check.
	lastDot := strings.LastIndex(state, ".")
	payload := state[:lastDot]
	parts := strings.SplitN(payload, "|", 2)
	nonce := parts[0]

	expiredExpiry := time.Now().Add(-time.Second).Unix()
	expiredPayload := nonce + "|" + itoa64(expiredExpiry)
	expiredSig := m.sign(expiredPayload)
	expiredState := expiredPayload + "." + expiredSig

	r := requestWithCookie(stateCookieName, expiredState)
	if err := m.ValidateState(expiredState, r); err != ErrInvalidState {
		t.Errorf("expected ErrInvalidState for expired state, got: %v", err)
	}
}

func TestValidateState_NoDelimiter(t *testing.T) {
	m := newTestManager(t)

	// A state string without the '.' delimiter is malformed.
	malformed := "nodotinhere"
	r := requestWithCookie(stateCookieName, malformed)
	if err := m.ValidateState(malformed, r); err != ErrInvalidState {
		t.Errorf("expected ErrInvalidState for malformed state, got: %v", err)
	}
}

func TestValidateState_DifferentManager(t *testing.T) {
	m1 := NewStateManager("secret-one-that-is-long-enough!!!!!", "localhost", false)
	m2 := NewStateManager("secret-two-that-is-long-enough!!!!!", "localhost", false)

	state, _, err := m1.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState: %v", err)
	}

	// m2 has a different secret; signature should fail.
	r := requestWithCookie(stateCookieName, state)
	if err := m2.ValidateState(state, r); err != ErrInvalidState {
		t.Errorf("expected ErrInvalidState when validating with different secret, got: %v", err)
	}
}

func TestClearStateCookie(t *testing.T) {
	m := NewStateManager("a-test-secret-that-is-long-enough!!", "example.com", true)
	cookie := m.ClearStateCookie()

	if cookie.Name != stateCookieName {
		t.Errorf("Name: got %q want %q", cookie.Name, stateCookieName)
	}
	if cookie.Value != "" {
		t.Errorf("Value: got %q want empty string", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Errorf("MaxAge: got %d want -1", cookie.MaxAge)
	}
	if !cookie.Secure {
		t.Error("expected Secure=true")
	}
	if !cookie.HttpOnly {
		t.Error("expected HttpOnly=true")
	}
	if cookie.Domain != "example.com" {
		t.Errorf("Domain: got %q want %q", cookie.Domain, "example.com")
	}
}

func TestGenerateState_Unique(t *testing.T) {
	m := newTestManager(t)
	const n = 20
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		state, _, err := m.GenerateState()
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if _, dup := seen[state]; dup {
			t.Fatalf("duplicate state at iteration %d", i)
		}
		seen[state] = struct{}{}
	}
}

// itoa64 converts an int64 to its decimal string representation.
func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
