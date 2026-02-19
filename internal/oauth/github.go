// Package oauth provides GitHub OAuth2 client functionality.
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GitHubUser represents a GitHub user profile.
type GitHubUser struct {
	ID        int64   `json:"id"`
	Login     string  `json:"login"`
	Email     string  `json:"email"`
	Name      *string `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}

// GitHubClient is an OAuth2 client for GitHub.
type GitHubClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// NewGitHubClient creates a new GitHubClient with the given credentials.
func NewGitHubClient(clientID, clientSecret string) *GitHubClient {
	return &GitHubClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{},
	}
}

// AuthorizeURL returns the GitHub OAuth authorization URL with scopes "read:user user:email".
func (c *GitHubClient) AuthorizeURL(redirectURI, state string) string {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "read:user user:email")
	params.Set("state", state)
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

// TokenResponse holds the response from GitHub's token endpoint.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// Exchange exchanges an authorization code for an access token.
func (c *GitHubClient) Exchange(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("client_secret", c.clientSecret)
	params.Set("code", code)
	params.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange returned status %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}

	return &token, nil
}

// GetUser retrieves the authenticated user's GitHub profile.
// If the user's email is private, it falls back to getPrimaryEmail.
func (c *GitHubClient) GetUser(ctx context.Context, accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("creating user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user returned status %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decoding user response: %w", err)
	}

	// If email is empty (private), fetch it separately.
	if user.Email == "" {
		email, err := c.getPrimaryEmail(ctx, accessToken)
		if err != nil {
			return nil, fmt.Errorf("fetching primary email: %w", err)
		}
		user.Email = email
	}

	return &user, nil
}

// githubEmail represents a single entry from the GitHub emails endpoint.
type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// getPrimaryEmail fetches the user's primary verified email from GitHub.
func (c *GitHubClient) getPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("creating emails request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching emails: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get emails returned status %d", resp.StatusCode)
	}

	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("decoding emails response: %w", err)
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	// Fall back to first verified email.
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}

	return "", fmt.Errorf("no verified email found")
}
