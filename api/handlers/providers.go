package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zoobzio/rocco"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/api/contracts"
	"github.com/zoobzio/sumatra/api/transformers"
	"github.com/zoobzio/sumatra/api/wire"
	"github.com/zoobzio/sumatra/config"
	intoauth "github.com/zoobzio/sumatra/internal/oauth"
	intsession "github.com/zoobzio/sumatra/internal/session"
	"github.com/zoobzio/sumatra/models"
)

// InitiateGitHubLink begins the GitHub OAuth flow for linking a provider to an existing account.
// The user must be authenticated. Generates a state cookie and redirects to GitHub.
var InitiateGitHubLink = rocco.GET("/providers/github/link", func(req *rocco.Request[rocco.NoBody]) (rocco.Redirect, error) {
	githubCfg := sum.MustUse[config.GitHub](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	stateMgr := intsession.NewStateManager(
		sessionCfg.StateSecret,
		sessionCfg.CookieDomain,
		sessionCfg.CookieSecure,
	)

	state, stateCookie, err := stateMgr.GenerateState()
	if err != nil {
		return rocco.Redirect{}, ErrGitHubOAuthFailed
	}

	client := intoauth.NewGitHubClient(githubCfg.ClientID, githubCfg.ClientSecret)
	authURL := client.AuthorizeURL(githubCfg.CallbackURL, state)

	headers := http.Header{}
	headers.Add("Set-Cookie", stateCookie.String())

	return rocco.Redirect{
		URL:     authURL,
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("Initiate GitHub link").
	WithDescription("Begins the GitHub OAuth flow for linking a GitHub account to the authenticated user.").
	WithTags("Providers").
	WithAuthentication().
	WithErrors(ErrGitHubOAuthFailed)

// GitHubLinkCallback completes the GitHub OAuth linking flow.
// Validates the state, exchanges the code, and creates a Provider record.
// If the GitHub account is already linked to another user, returns an error.
var GitHubLinkCallback = rocco.GET("/providers/github/callback", func(req *rocco.Request[rocco.NoBody]) (rocco.Redirect, error) {
	providers := sum.MustUse[contracts.Providers](req.Context)
	githubCfg := sum.MustUse[config.GitHub](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	stateMgr := intsession.NewStateManager(
		sessionCfg.StateSecret,
		sessionCfg.CookieDomain,
		sessionCfg.CookieSecure,
	)

	stateParam := req.Params.Query["state"]
	code := req.Params.Query["code"]

	if err := stateMgr.ValidateState(stateParam, req.Request); err != nil {
		headers := http.Header{}
		headers.Add("Set-Cookie", stateMgr.ClearStateCookie().String())
		return rocco.Redirect{URL: "/?error=invalid_state", Status: http.StatusFound, Headers: headers}, nil
	}

	headers := http.Header{}
	headers.Add("Set-Cookie", stateMgr.ClearStateCookie().String())

	client := intoauth.NewGitHubClient(githubCfg.ClientID, githubCfg.ClientSecret)

	token, err := client.Exchange(req.Context, code, githubCfg.CallbackURL)
	if err != nil {
		return rocco.Redirect{URL: "/?error=oauth_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	ghUser, err := client.GetUser(req.Context, token.AccessToken)
	if err != nil {
		return rocco.Redirect{URL: "/?error=oauth_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	providerUserID := fmt.Sprintf("%d", ghUser.ID)

	// Check whether this GitHub account is already linked to a different user.
	existing, err := providers.GetByProviderUser(req.Context, models.ProviderTypeGitHub, providerUserID)
	if err == nil && existing != nil && existing.UserID != req.Identity.ID() {
		return rocco.Redirect{URL: "/?error=provider_already_linked", Status: http.StatusFound, Headers: headers}, nil
	}

	// Create or update the provider record for the current user.
	now := time.Now()
	provider := &models.Provider{
		UserID:         req.Identity.ID(),
		Type:           models.ProviderTypeGitHub,
		ProviderUserID: providerUserID,
		AccessToken:    token.AccessToken,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := providers.Set(req.Context, "", provider); err != nil {
		return rocco.Redirect{URL: "/?error=link_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	return rocco.Redirect{
		URL:     "/?linked=github",
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("GitHub link callback").
	WithDescription("Completes the GitHub OAuth linking flow. Links the GitHub account to the authenticated user.").
	WithTags("Providers").
	WithQueryParams("code", "state").
	WithAuthentication()

// UnlinkGitHub removes the GitHub provider link for the authenticated user.
// The user must have at least one other authentication method (password or another provider).
var UnlinkGitHub = rocco.DELETE("/providers/github", func(req *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	users := sum.MustUse[contracts.Users](req.Context)
	providers := sum.MustUse[contracts.Providers](req.Context)

	// Ensure the provider link exists for this user.
	_, err := providers.GetByUserAndType(req.Context, req.Identity.ID(), models.ProviderTypeGitHub)
	if err != nil {
		return rocco.NoBody{}, ErrProviderNotFound
	}

	// Verify the user retains at least one other authentication method.
	user, err := users.Get(req.Context, req.Identity.ID())
	if err != nil || user == nil {
		return rocco.NoBody{}, ErrUserNotFound
	}

	allProviders, err := providers.ListByUser(req.Context, req.Identity.ID())
	if err != nil {
		return rocco.NoBody{}, ErrProviderLinkFailed
	}

	// Count remaining auth methods: password + every other linked provider.
	remainingMethods := 0
	if user.PasswordHash != nil {
		remainingMethods++
	}
	for _, p := range allProviders {
		if p.Type != models.ProviderTypeGitHub {
			remainingMethods++
		}
	}
	if remainingMethods == 0 {
		return rocco.NoBody{}, ErrLastAuthMethod
	}

	if err := providers.DeleteByUserAndType(req.Context, req.Identity.ID(), models.ProviderTypeGitHub); err != nil {
		return rocco.NoBody{}, ErrProviderLinkFailed
	}

	return rocco.NoBody{}, nil
}).WithSummary("Unlink GitHub").
	WithDescription("Removes the GitHub provider link for the authenticated user. Requires at least one other authentication method to remain.").
	WithTags("Providers").
	WithAuthentication().
	WithSuccessStatus(204).
	WithErrors(ErrProviderNotFound, ErrUserNotFound, ErrLastAuthMethod, ErrProviderLinkFailed)

// ListProviders returns all linked OAuth providers for the authenticated user.
var ListProviders = rocco.GET("/providers", func(req *rocco.Request[rocco.NoBody]) (wire.ProviderListResponse, error) {
	providers := sum.MustUse[contracts.Providers](req.Context)

	list, err := providers.ListByUser(req.Context, req.Identity.ID())
	if err != nil {
		return wire.ProviderListResponse{}, ErrProviderLinkFailed
	}

	return transformers.ProvidersToList(list), nil
}).WithSummary("List linked providers").
	WithDescription("Returns all linked OAuth providers for the authenticated user.").
	WithTags("Providers").
	WithAuthentication().
	WithErrors(ErrProviderLinkFailed)

// InitiateGitHubLogin begins the GitHub OAuth flow for logging in via a linked GitHub account.
// No authentication is required — this is a login entry point.
var InitiateGitHubLogin = rocco.GET("/login/github", func(req *rocco.Request[rocco.NoBody]) (rocco.Redirect, error) {
	githubCfg := sum.MustUse[config.GitHub](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	stateMgr := intsession.NewStateManager(
		sessionCfg.StateSecret,
		sessionCfg.CookieDomain,
		sessionCfg.CookieSecure,
	)

	state, stateCookie, err := stateMgr.GenerateState()
	if err != nil {
		return rocco.Redirect{}, ErrGitHubOAuthFailed
	}

	client := intoauth.NewGitHubClient(githubCfg.ClientID, githubCfg.ClientSecret)
	authURL := client.AuthorizeURL(githubCfg.CallbackURL, state)

	headers := http.Header{}
	headers.Add("Set-Cookie", stateCookie.String())

	return rocco.Redirect{
		URL:     authURL,
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("Login via GitHub").
	WithDescription("Initiates the GitHub OAuth flow for logging in via a linked GitHub account.").
	WithTags("Auth").
	WithErrors(ErrGitHubOAuthFailed)

// GitHubLoginCallback completes the GitHub OAuth login flow.
// Validates the state, exchanges the code, finds the linked Provider, and creates a session.
var GitHubLoginCallback = rocco.GET("/login/github/callback", func(req *rocco.Request[rocco.NoBody]) (rocco.Redirect, error) {
	providers := sum.MustUse[contracts.Providers](req.Context)
	sessions := sum.MustUse[contracts.Sessions](req.Context)
	githubCfg := sum.MustUse[config.GitHub](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	stateMgr := intsession.NewStateManager(
		sessionCfg.StateSecret,
		sessionCfg.CookieDomain,
		sessionCfg.CookieSecure,
	)

	stateParam := req.Params.Query["state"]
	code := req.Params.Query["code"]

	if err := stateMgr.ValidateState(stateParam, req.Request); err != nil {
		headers := http.Header{}
		headers.Add("Set-Cookie", stateMgr.ClearStateCookie().String())
		return rocco.Redirect{URL: "/login?error=invalid_state", Status: http.StatusFound, Headers: headers}, nil
	}

	headers := http.Header{}
	headers.Add("Set-Cookie", stateMgr.ClearStateCookie().String())

	client := intoauth.NewGitHubClient(githubCfg.ClientID, githubCfg.ClientSecret)

	token, err := client.Exchange(req.Context, code, githubCfg.CallbackURL)
	if err != nil {
		return rocco.Redirect{URL: "/login?error=oauth_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	ghUser, err := client.GetUser(req.Context, token.AccessToken)
	if err != nil {
		return rocco.Redirect{URL: "/login?error=oauth_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	providerUserID := fmt.Sprintf("%d", ghUser.ID)

	// Find the account linked to this GitHub identity.
	provider, err := providers.GetByProviderUser(req.Context, models.ProviderTypeGitHub, providerUserID)
	if err != nil || provider == nil {
		return rocco.Redirect{URL: "/login?error=account_not_linked", Status: http.StatusFound, Headers: headers}, nil
	}

	// Create a session for the linked user.
	sessionToken, err := intsession.GenerateToken()
	if err != nil {
		return rocco.Redirect{URL: "/login?error=login_failed", Status: http.StatusFound, Headers: headers}, nil
	}
	now := time.Now()
	sess := &models.Session{
		Token:     sessionToken,
		UserID:    provider.UserID,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionCfg.TTL),
	}
	if err := sessions.SetWithUserIndex(req.Context, sess, sessionCfg.TTL); err != nil {
		return rocco.Redirect{URL: "/login?error=login_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	headers.Add("Set-Cookie", buildSessionCookie(sessionCfg, sessionToken).String())

	return rocco.Redirect{
		URL:     "/",
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("GitHub login callback").
	WithDescription("Completes the GitHub OAuth login flow. Finds the linked account and creates a session.").
	WithTags("Auth").
	WithQueryParams("code", "state")

// =============================================================================
// Google OAuth Handlers
// =============================================================================

// InitiateGoogleLink begins the Google OAuth flow for linking a provider to an existing account.
// The user must be authenticated. Generates a state cookie and redirects to Google.
var InitiateGoogleLink = rocco.GET("/providers/google/link", func(req *rocco.Request[rocco.NoBody]) (rocco.Redirect, error) {
	googleCfg := sum.MustUse[config.Google](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	stateMgr := intsession.NewStateManager(
		sessionCfg.StateSecret,
		sessionCfg.CookieDomain,
		sessionCfg.CookieSecure,
	)

	state, stateCookie, err := stateMgr.GenerateState()
	if err != nil {
		return rocco.Redirect{}, ErrGoogleOAuthFailed
	}

	client := intoauth.NewGoogleClient(googleCfg.ClientID, googleCfg.ClientSecret)
	authURL := client.AuthorizeURL(googleCfg.CallbackURL, state)

	headers := http.Header{}
	headers.Add("Set-Cookie", stateCookie.String())

	return rocco.Redirect{
		URL:     authURL,
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("Initiate Google link").
	WithDescription("Begins the Google OAuth flow for linking a Google account to the authenticated user.").
	WithTags("Providers").
	WithAuthentication().
	WithErrors(ErrGoogleOAuthFailed)

// GoogleLinkCallback completes the Google OAuth linking flow.
// Validates the state, exchanges the code, and creates a Provider record.
// If the Google account is already linked to another user, returns an error.
var GoogleLinkCallback = rocco.GET("/providers/google/callback", func(req *rocco.Request[rocco.NoBody]) (rocco.Redirect, error) {
	providers := sum.MustUse[contracts.Providers](req.Context)
	googleCfg := sum.MustUse[config.Google](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	stateMgr := intsession.NewStateManager(
		sessionCfg.StateSecret,
		sessionCfg.CookieDomain,
		sessionCfg.CookieSecure,
	)

	stateParam := req.Params.Query["state"]
	code := req.Params.Query["code"]

	if err := stateMgr.ValidateState(stateParam, req.Request); err != nil {
		headers := http.Header{}
		headers.Add("Set-Cookie", stateMgr.ClearStateCookie().String())
		return rocco.Redirect{URL: "/?error=invalid_state", Status: http.StatusFound, Headers: headers}, nil
	}

	headers := http.Header{}
	headers.Add("Set-Cookie", stateMgr.ClearStateCookie().String())

	client := intoauth.NewGoogleClient(googleCfg.ClientID, googleCfg.ClientSecret)

	token, err := client.Exchange(req.Context, code, googleCfg.CallbackURL)
	if err != nil {
		return rocco.Redirect{URL: "/?error=oauth_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	googleUser, err := client.GetUser(req.Context, token.AccessToken)
	if err != nil {
		return rocco.Redirect{URL: "/?error=oauth_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	// Check whether this Google account is already linked to a different user.
	existing, err := providers.GetByProviderUser(req.Context, models.ProviderTypeGoogle, googleUser.ID)
	if err == nil && existing != nil && existing.UserID != req.Identity.ID() {
		return rocco.Redirect{URL: "/?error=provider_already_linked", Status: http.StatusFound, Headers: headers}, nil
	}

	// Create or update the provider record for the current user.
	now := time.Now()
	provider := &models.Provider{
		UserID:         req.Identity.ID(),
		Type:           models.ProviderTypeGoogle,
		ProviderUserID: googleUser.ID,
		AccessToken:    token.AccessToken,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := providers.Set(req.Context, "", provider); err != nil {
		return rocco.Redirect{URL: "/?error=link_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	return rocco.Redirect{
		URL:     "/?linked=google",
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("Google link callback").
	WithDescription("Completes the Google OAuth linking flow. Links the Google account to the authenticated user.").
	WithTags("Providers").
	WithQueryParams("code", "state").
	WithAuthentication()

// UnlinkGoogle removes the Google provider link for the authenticated user.
// The user must have at least one other authentication method (password or another provider).
var UnlinkGoogle = rocco.DELETE("/providers/google", func(req *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	users := sum.MustUse[contracts.Users](req.Context)
	providers := sum.MustUse[contracts.Providers](req.Context)

	// Ensure the provider link exists for this user.
	_, err := providers.GetByUserAndType(req.Context, req.Identity.ID(), models.ProviderTypeGoogle)
	if err != nil {
		return rocco.NoBody{}, ErrProviderNotFound
	}

	// Verify the user retains at least one other authentication method.
	user, err := users.Get(req.Context, req.Identity.ID())
	if err != nil || user == nil {
		return rocco.NoBody{}, ErrUserNotFound
	}

	allProviders, err := providers.ListByUser(req.Context, req.Identity.ID())
	if err != nil {
		return rocco.NoBody{}, ErrProviderLinkFailed
	}

	// Count remaining auth methods: password + every other linked provider.
	remainingMethods := 0
	if user.PasswordHash != nil {
		remainingMethods++
	}
	for _, p := range allProviders {
		if p.Type != models.ProviderTypeGoogle {
			remainingMethods++
		}
	}
	if remainingMethods == 0 {
		return rocco.NoBody{}, ErrLastAuthMethod
	}

	if err := providers.DeleteByUserAndType(req.Context, req.Identity.ID(), models.ProviderTypeGoogle); err != nil {
		return rocco.NoBody{}, ErrProviderLinkFailed
	}

	return rocco.NoBody{}, nil
}).WithSummary("Unlink Google").
	WithDescription("Removes the Google provider link for the authenticated user. Requires at least one other authentication method to remain.").
	WithTags("Providers").
	WithAuthentication().
	WithSuccessStatus(204).
	WithErrors(ErrProviderNotFound, ErrUserNotFound, ErrLastAuthMethod, ErrProviderLinkFailed)

// InitiateGoogleLogin begins the Google OAuth flow for logging in via a linked Google account.
// No authentication is required — this is a login entry point.
var InitiateGoogleLogin = rocco.GET("/login/google", func(req *rocco.Request[rocco.NoBody]) (rocco.Redirect, error) {
	googleCfg := sum.MustUse[config.Google](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	stateMgr := intsession.NewStateManager(
		sessionCfg.StateSecret,
		sessionCfg.CookieDomain,
		sessionCfg.CookieSecure,
	)

	state, stateCookie, err := stateMgr.GenerateState()
	if err != nil {
		return rocco.Redirect{}, ErrGoogleOAuthFailed
	}

	client := intoauth.NewGoogleClient(googleCfg.ClientID, googleCfg.ClientSecret)
	authURL := client.AuthorizeURL(googleCfg.CallbackURL, state)

	headers := http.Header{}
	headers.Add("Set-Cookie", stateCookie.String())

	return rocco.Redirect{
		URL:     authURL,
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("Login via Google").
	WithDescription("Initiates the Google OAuth flow for logging in via a linked Google account.").
	WithTags("Auth").
	WithErrors(ErrGoogleOAuthFailed)

// GoogleLoginCallback completes the Google OAuth login flow.
// Validates the state, exchanges the code, finds the linked Provider, and creates a session.
var GoogleLoginCallback = rocco.GET("/login/google/callback", func(req *rocco.Request[rocco.NoBody]) (rocco.Redirect, error) {
	providers := sum.MustUse[contracts.Providers](req.Context)
	sessions := sum.MustUse[contracts.Sessions](req.Context)
	googleCfg := sum.MustUse[config.Google](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	stateMgr := intsession.NewStateManager(
		sessionCfg.StateSecret,
		sessionCfg.CookieDomain,
		sessionCfg.CookieSecure,
	)

	stateParam := req.Params.Query["state"]
	code := req.Params.Query["code"]

	if err := stateMgr.ValidateState(stateParam, req.Request); err != nil {
		headers := http.Header{}
		headers.Add("Set-Cookie", stateMgr.ClearStateCookie().String())
		return rocco.Redirect{URL: "/login?error=invalid_state", Status: http.StatusFound, Headers: headers}, nil
	}

	headers := http.Header{}
	headers.Add("Set-Cookie", stateMgr.ClearStateCookie().String())

	client := intoauth.NewGoogleClient(googleCfg.ClientID, googleCfg.ClientSecret)

	token, err := client.Exchange(req.Context, code, googleCfg.CallbackURL)
	if err != nil {
		return rocco.Redirect{URL: "/login?error=oauth_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	googleUser, err := client.GetUser(req.Context, token.AccessToken)
	if err != nil {
		return rocco.Redirect{URL: "/login?error=oauth_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	// Find the account linked to this Google identity.
	provider, err := providers.GetByProviderUser(req.Context, models.ProviderTypeGoogle, googleUser.ID)
	if err != nil || provider == nil {
		return rocco.Redirect{URL: "/login?error=account_not_linked", Status: http.StatusFound, Headers: headers}, nil
	}

	// Create a session for the linked user.
	sessionToken, err := intsession.GenerateToken()
	if err != nil {
		return rocco.Redirect{URL: "/login?error=login_failed", Status: http.StatusFound, Headers: headers}, nil
	}
	now := time.Now()
	sess := &models.Session{
		Token:     sessionToken,
		UserID:    provider.UserID,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionCfg.TTL),
	}
	if err := sessions.SetWithUserIndex(req.Context, sess, sessionCfg.TTL); err != nil {
		return rocco.Redirect{URL: "/login?error=login_failed", Status: http.StatusFound, Headers: headers}, nil
	}

	headers.Add("Set-Cookie", buildSessionCookie(sessionCfg, sessionToken).String())

	return rocco.Redirect{
		URL:     "/",
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("Google login callback").
	WithDescription("Completes the Google OAuth login flow. Finds the linked account and creates a session.").
	WithTags("Auth").
	WithQueryParams("code", "state")
