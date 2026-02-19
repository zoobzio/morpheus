package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GoogleUser represents a Google user profile from the userinfo endpoint.
type GoogleUser struct {
	ID            string  `json:"sub"`
	Email         string  `json:"email"`
	EmailVerified bool    `json:"email_verified"`
	Name          *string `json:"name"`
	Picture       *string `json:"picture"`
}

// GoogleClient is an OAuth2 client for Google.
type GoogleClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// NewGoogleClient creates a new GoogleClient with the given credentials.
func NewGoogleClient(clientID, clientSecret string) *GoogleClient {
	return &GoogleClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{},
	}
}

// AuthorizeURL returns the Google OAuth authorization URL with scopes for email and profile.
func (c *GoogleClient) AuthorizeURL(redirectURI, state string) string {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("state", state)
	params.Set("access_type", "offline")
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

// GoogleTokenResponse holds the response from Google's token endpoint.
type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token,omitempty"`
}

// Exchange exchanges an authorization code for an access token.
func (c *GoogleClient) Exchange(ctx context.Context, code, redirectURI string) (*GoogleTokenResponse, error) {
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("client_secret", c.clientSecret)
	params.Set("code", code)
	params.Set("redirect_uri", redirectURI)
	params.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(params.Encode()))
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

	var token GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}

	return &token, nil
}

// GetUser retrieves the authenticated user's Google profile using the userinfo endpoint.
func (c *GoogleClient) GetUser(ctx context.Context, accessToken string) (*GoogleUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("creating userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching userinfo: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get userinfo returned status %d", resp.StatusCode)
	}

	var user GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decoding userinfo response: %w", err)
	}

	if !user.EmailVerified {
		return nil, fmt.Errorf("email not verified")
	}

	return &user, nil
}
