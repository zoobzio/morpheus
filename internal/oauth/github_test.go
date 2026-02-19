package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestClient(serverURL string) *GitHubClient {
	return &GitHubClient{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		httpClient:   &http.Client{},
	}
}

// withBaseURL returns a shallow copy of the client whose httpClient
// always rewrites the host to the given server.
func newTestClientWithServer(srv *httptest.Server) *GitHubClient {
	return &GitHubClient{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		httpClient:   srv.Client(),
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// AuthorizeURL
// ──────────────────────────────────────────────────────────────────────────────

func TestAuthorizeURL_ContainsClientID(t *testing.T) {
	c := newTestClient("")
	u := c.AuthorizeURL("https://example.com/callback", "mystate")
	if !strings.Contains(u, "client_id=test-client-id") {
		t.Errorf("expected client_id in URL: %q", u)
	}
}

func TestAuthorizeURL_ContainsRedirectURI(t *testing.T) {
	c := newTestClient("")
	redirect := "https://example.com/callback"
	u := c.AuthorizeURL(redirect, "mystate")
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}
	got := parsed.Query().Get("redirect_uri")
	if got != redirect {
		t.Errorf("redirect_uri: got %q want %q", got, redirect)
	}
}

func TestAuthorizeURL_ContainsState(t *testing.T) {
	c := newTestClient("")
	u := c.AuthorizeURL("https://example.com/callback", "mystate")
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}
	got := parsed.Query().Get("state")
	if got != "mystate" {
		t.Errorf("state: got %q want %q", got, "mystate")
	}
}

func TestAuthorizeURL_ContainsScopes(t *testing.T) {
	c := newTestClient("")
	u := c.AuthorizeURL("https://example.com/callback", "mystate")
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}
	scope := parsed.Query().Get("scope")
	if !strings.Contains(scope, "read:user") {
		t.Errorf("scope missing read:user: %q", scope)
	}
	if !strings.Contains(scope, "user:email") {
		t.Errorf("scope missing user:email: %q", scope)
	}
}

func TestAuthorizeURL_BaseURL(t *testing.T) {
	c := newTestClient("")
	u := c.AuthorizeURL("https://example.com/callback", "mystate")
	if !strings.HasPrefix(u, "https://github.com/login/oauth/authorize") {
		t.Errorf("unexpected base URL: %q", u)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Exchange
// ──────────────────────────────────────────────────────────────────────────────

func TestExchange_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Exchange: expected POST, got %s", r.Method)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Exchange: missing Accept: application/json header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "gho_test_token",
			TokenType:   "bearer",
			Scope:       "read:user,user:email",
		})
	}))
	defer srv.Close()

	c := &GitHubClient{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		httpClient:   srv.Client(),
	}
	// Redirect the token endpoint to the test server.
	// Because the client always hits the hardcoded URL, we use a transport
	// that rewrites the host.
	c.httpClient.Transport = rewriteTransport(srv.URL, c.httpClient.Transport)

	tok, err := c.Exchange(context.Background(), "auth-code", "https://example.com/callback")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tok.AccessToken != "gho_test_token" {
		t.Errorf("AccessToken: got %q want %q", tok.AccessToken, "gho_test_token")
	}
}

func TestExchange_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := &GitHubClient{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		httpClient:   srv.Client(),
	}
	c.httpClient.Transport = rewriteTransport(srv.URL, c.httpClient.Transport)

	_, err := c.Exchange(context.Background(), "bad-code", "https://example.com/callback")
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}

func TestExchange_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := &GitHubClient{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		httpClient:   srv.Client(),
	}
	c.httpClient.Transport = rewriteTransport(srv.URL, c.httpClient.Transport)

	_, err := c.Exchange(context.Background(), "code", "https://example.com/callback")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// GetUser — email present in profile
// ──────────────────────────────────────────────────────────────────────────────

func TestGetUser_WithEmail(t *testing.T) {
	name := "The Octocat"
	avatar := "https://avatars.githubusercontent.com/u/583231"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Errorf("unexpected path: %q", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer gho_test_token" {
			t.Errorf("Authorization: got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GitHubUser{
			ID:        583231,
			Login:     "octocat",
			Email:     "octocat@github.com",
			Name:      &name,
			AvatarURL: &avatar,
		})
	}))
	defer srv.Close()

	c := &GitHubClient{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		httpClient:   srv.Client(),
	}
	c.httpClient.Transport = rewriteTransport(srv.URL, c.httpClient.Transport)

	user, err := c.GetUser(context.Background(), "gho_test_token")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.Login != "octocat" {
		t.Errorf("Login: got %q want %q", user.Login, "octocat")
	}
	if user.Email != "octocat@github.com" {
		t.Errorf("Email: got %q want %q", user.Email, "octocat@github.com")
	}
	if user.ID != 583231 {
		t.Errorf("ID: got %d want %d", user.ID, 583231)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// GetUser — email absent; falls back to /user/emails
// ──────────────────────────────────────────────────────────────────────────────

func TestGetUser_FallbackToEmailEndpoint(t *testing.T) {
	name := "The Octocat"
	avatar := "https://avatars.githubusercontent.com/u/583231"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/user":
			// Return user without email (private).
			_ = json.NewEncoder(w).Encode(GitHubUser{
				ID:        583231,
				Login:     "octocat",
				Email:     "",
				Name:      &name,
				AvatarURL: &avatar,
			})
		case "/user/emails":
			_ = json.NewEncoder(w).Encode([]githubEmail{
				{Email: "noreply@users.noreply.github.com", Primary: false, Verified: true},
				{Email: "octocat@github.com", Primary: true, Verified: true},
			})
		default:
			t.Errorf("unexpected path: %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := &GitHubClient{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		httpClient:   srv.Client(),
	}
	c.httpClient.Transport = rewriteTransport(srv.URL, c.httpClient.Transport)

	user, err := c.GetUser(context.Background(), "gho_test_token")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.Email != "octocat@github.com" {
		t.Errorf("Email: got %q want %q", user.Email, "octocat@github.com")
	}
}

func TestGetUser_FallbackEmailFallsBackToFirstVerified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/user":
			_ = json.NewEncoder(w).Encode(GitHubUser{ID: 1, Login: "user1"})
		case "/user/emails":
			// No primary — should fall back to first verified.
			_ = json.NewEncoder(w).Encode([]githubEmail{
				{Email: "first@example.com", Primary: false, Verified: true},
				{Email: "second@example.com", Primary: false, Verified: true},
			})
		}
	}))
	defer srv.Close()

	c := &GitHubClient{
		clientID:   "id",
		clientSecret: "secret",
		httpClient: srv.Client(),
	}
	c.httpClient.Transport = rewriteTransport(srv.URL, c.httpClient.Transport)

	user, err := c.GetUser(context.Background(), "tok")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.Email != "first@example.com" {
		t.Errorf("Email: got %q want %q", user.Email, "first@example.com")
	}
}

func TestGetUser_NoVerifiedEmail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/user":
			_ = json.NewEncoder(w).Encode(GitHubUser{ID: 1, Login: "user1"})
		case "/user/emails":
			_ = json.NewEncoder(w).Encode([]githubEmail{
				{Email: "unverified@example.com", Primary: false, Verified: false},
			})
		}
	}))
	defer srv.Close()

	c := &GitHubClient{
		clientID:   "id",
		clientSecret: "secret",
		httpClient: srv.Client(),
	}
	c.httpClient.Transport = rewriteTransport(srv.URL, c.httpClient.Transport)

	_, err := c.GetUser(context.Background(), "tok")
	if err == nil {
		t.Fatal("expected error for no verified email, got nil")
	}
}

func TestGetUser_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := &GitHubClient{
		clientID:   "id",
		clientSecret: "secret",
		httpClient: srv.Client(),
	}
	c.httpClient.Transport = rewriteTransport(srv.URL, c.httpClient.Transport)

	_, err := c.GetUser(context.Background(), "bad-token")
	if err == nil {
		t.Fatal("expected error for non-200 user response, got nil")
	}
}

func TestGetUser_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := &GitHubClient{
		clientID:   "id",
		clientSecret: "secret",
		httpClient: srv.Client(),
	}
	c.httpClient.Transport = rewriteTransport(srv.URL, c.httpClient.Transport)

	_, err := c.GetUser(context.Background(), "tok")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// helpers
// ──────────────────────────────────────────────────────────────────────────────

// rewriteTransport wraps an existing RoundTripper and rewrites the scheme+host
// of every request to the given base URL.  This lets us point the hardcoded
// GitHub URLs at a local httptest.Server without modifying production code.
type rewriteRoundTripper struct {
	base    *url.URL
	wrapped http.RoundTripper
}

func rewriteTransport(base string, wrapped http.RoundTripper) http.RoundTripper {
	u, _ := url.Parse(base)
	if wrapped == nil {
		wrapped = http.DefaultTransport
	}
	return &rewriteRoundTripper{base: u, wrapped: wrapped}
}

func (t *rewriteRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	r2.URL.Scheme = t.base.Scheme
	r2.URL.Host = t.base.Host
	return t.wrapped.RoundTrip(r2)
}
