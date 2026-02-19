package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	stateCookieName = "oauth_state"
	stateTTL        = 10 * time.Minute
)

// ErrInvalidState is returned when the OAuth state parameter fails validation.
var ErrInvalidState = errors.New("invalid oauth state")

// StateManager handles generation and validation of OAuth state parameters.
type StateManager struct {
	secret       []byte
	cookieDomain string
	cookieSecure bool
}

// NewStateManager creates a new StateManager with the given HMAC secret and cookie settings.
func NewStateManager(secret, cookieDomain string, cookieSecure bool) *StateManager {
	return &StateManager{
		secret:       []byte(secret),
		cookieDomain: cookieDomain,
		cookieSecure: cookieSecure,
	}
}

// GenerateState creates a signed state string and its corresponding cookie.
// The state encodes a nonce and expiry, signed with HMAC-SHA256.
func (m *StateManager) GenerateState() (string, *http.Cookie, error) {
	nonce, err := GenerateToken()
	if err != nil {
		return "", nil, fmt.Errorf("generating nonce: %w", err)
	}

	expiry := time.Now().Add(stateTTL).Unix()
	payload := nonce + "|" + strconv.FormatInt(expiry, 10)

	sig := m.sign(payload)
	state := payload + "." + sig

	cookie := &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		Domain:   m.cookieDomain,
		MaxAge:   int(stateTTL.Seconds()),
		Secure:   m.cookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	return state, cookie, nil
}

// ValidateState verifies the state parameter against the cookie stored in the request.
// It checks that the state matches the cookie, verifies the HMAC signature,
// and ensures the state has not expired.
func (m *StateManager) ValidateState(stateParam string, r *http.Request) error {
	cookie, err := r.Cookie(stateCookieName)
	if err != nil {
		return ErrInvalidState
	}

	if cookie.Value != stateParam {
		return ErrInvalidState
	}

	// Split into payload and signature.
	lastDot := strings.LastIndex(stateParam, ".")
	if lastDot < 0 {
		return ErrInvalidState
	}
	payload := stateParam[:lastDot]
	sig := stateParam[lastDot+1:]

	if !hmac.Equal([]byte(sig), []byte(m.sign(payload))) {
		return ErrInvalidState
	}

	// Verify expiry from payload: nonce|expiry
	parts := strings.SplitN(payload, "|", 2)
	if len(parts) != 2 {
		return ErrInvalidState
	}
	expiry, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return ErrInvalidState
	}
	if time.Now().Unix() > expiry {
		return ErrInvalidState
	}

	return nil
}

// ClearStateCookie returns an expired cookie that clears the state cookie in the browser.
func (m *StateManager) ClearStateCookie() *http.Cookie {
	return &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/",
		Domain:   m.cookieDomain,
		MaxAge:   -1,
		Secure:   m.cookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// sign returns a base64url-encoded HMAC-SHA256 signature of the payload.
func (m *StateManager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
