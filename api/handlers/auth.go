package handlers

import (
	"net/http"
	"time"

	"github.com/zoobzio/rocco"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/api/contracts"
	"github.com/zoobzio/sumatra/api/transformers"
	"github.com/zoobzio/sumatra/api/wire"
	"github.com/zoobzio/sumatra/config"
	extpostmark "github.com/zoobzio/sumatra/external/postmark"
	intpassword "github.com/zoobzio/sumatra/internal/password"
	intsession "github.com/zoobzio/sumatra/internal/session"
	"github.com/zoobzio/sumatra/models"
)

// buildSessionCookie constructs the session cookie from config and token.
func buildSessionCookie(cfg config.Session, token string) *http.Cookie {
	return &http.Cookie{
		Name:     cfg.CookieName,
		Value:    token,
		Path:     cfg.CookiePath,
		Domain:   cfg.CookieDomain,
		MaxAge:   int(cfg.TTL.Seconds()),
		Secure:   cfg.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// Register creates a new user account.
// The user must verify their email before they can log in.
var Register = rocco.POST("/register", func(req *rocco.Request[wire.RegisterRequest]) (wire.UserResponse, error) {
	users := sum.MustUse[contracts.Users](req.Context)
	verificationTokens := sum.MustUse[contracts.VerificationTokens](req.Context)
	tokensCfg := sum.MustUse[config.Tokens](req.Context)
	postmarkCfg := sum.MustUse[config.Postmark](req.Context)

	// Reject if email is already registered.
	existing, err := users.GetByEmail(req.Context, req.Body.Email)
	if err == nil && existing != nil {
		return wire.UserResponse{}, ErrEmailAlreadyExists
	}

	// Hash the password.
	hash, err := intpassword.Hash(req.Body.Password)
	if err != nil {
		return wire.UserResponse{}, ErrRegistrationFailed
	}

	// Generate a token-based ID for the new user.
	userID, err := intsession.GenerateToken()
	if err != nil {
		return wire.UserResponse{}, ErrRegistrationFailed
	}

	// Create the user.
	user := &models.User{
		ID:            userID,
		Email:         req.Body.Email,
		PasswordHash:  &hash,
		EmailVerified: false,
	}
	if err := users.Set(req.Context, user.ID, user); err != nil {
		return wire.UserResponse{}, ErrRegistrationFailed
	}

	// Generate email verification token.
	rawToken, err := intsession.GenerateToken()
	if err != nil {
		return wire.UserResponse{}, ErrRegistrationFailed
	}
	now := time.Now()
	vt := &models.VerificationToken{
		Token:     rawToken,
		UserID:    user.ID,
		Type:      models.TokenTypeEmailVerify,
		CreatedAt: now,
		ExpiresAt: now.Add(tokensCfg.EmailVerifyTTL),
	}
	if err := verificationTokens.Set(req.Context, vt, tokensCfg.EmailVerifyTTL); err != nil {
		return wire.UserResponse{}, ErrRegistrationFailed
	}

	// Send verification email (best-effort; don't fail registration on send error).
	emailClient := extpostmark.NewClient(postmarkCfg.ServerToken, postmarkCfg.DefaultFrom)
	_, _ = emailClient.SendEmail(req.Context, extpostmark.EmailRequest{
		To:       user.Email,
		Subject:  "Verify your email address",
		TextBody: "Please verify your email address by submitting the following token:\n\n" + rawToken + "\n\nThis token expires in 24 hours.",
	})

	return transformers.UserToResponse(user), nil
}).WithSummary("Register").
	WithDescription("Creates a new user account. The user must verify their email before logging in.").
	WithTags("Auth").
	WithSuccessStatus(201).
	WithErrors(ErrEmailAlreadyExists, ErrRegistrationFailed)

// Login authenticates a user with email and password.
// The user's email must be verified. On success, redirects to / with a session cookie.
var Login = rocco.POST("/login", func(req *rocco.Request[wire.LoginRequest]) (rocco.Redirect, error) {
	users := sum.MustUse[contracts.Users](req.Context)
	sessions := sum.MustUse[contracts.Sessions](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	// Find user by email.
	user, err := users.GetByEmail(req.Context, req.Body.Email)
	if err != nil || user == nil {
		return rocco.Redirect{}, ErrInvalidCredentials
	}

	// Require a stored password hash.
	if user.PasswordHash == nil {
		return rocco.Redirect{}, ErrInvalidCredentials
	}

	// Verify password.
	ok, err := intpassword.Verify(req.Body.Password, *user.PasswordHash)
	if err != nil || !ok {
		return rocco.Redirect{}, ErrInvalidCredentials
	}

	// Require verified email.
	if !user.EmailVerified {
		return rocco.Redirect{}, ErrEmailNotVerified
	}

	// Create session.
	sessionToken, err := intsession.GenerateToken()
	if err != nil {
		return rocco.Redirect{}, ErrLoginFailed
	}
	now := time.Now()
	sess := &models.Session{
		Token:     sessionToken,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionCfg.TTL),
	}
	if err := sessions.SetWithUserIndex(req.Context, sess, sessionCfg.TTL); err != nil {
		return rocco.Redirect{}, ErrLoginFailed
	}

	headers := http.Header{}
	headers.Add("Set-Cookie", buildSessionCookie(sessionCfg, sessionToken).String())

	return rocco.Redirect{
		URL:     "/",
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("Login").
	WithDescription("Authenticates a user with email and password. Redirects with session cookie on success.").
	WithTags("Auth").
	WithErrors(ErrInvalidCredentials, ErrEmailNotVerified, ErrLoginFailed)

// RequestMagicLink sends a magic link sign-in email to the user.
// Always responds 204 so callers cannot enumerate registered emails.
var RequestMagicLink = rocco.POST("/login/magic", func(req *rocco.Request[wire.MagicLinkRequest]) (rocco.NoBody, error) {
	users := sum.MustUse[contracts.Users](req.Context)
	verificationTokens := sum.MustUse[contracts.VerificationTokens](req.Context)
	tokensCfg := sum.MustUse[config.Tokens](req.Context)
	postmarkCfg := sum.MustUse[config.Postmark](req.Context)

	user, err := users.GetByEmail(req.Context, req.Body.Email)
	if err != nil || user == nil || !user.EmailVerified {
		// Do not reveal whether the email exists or is unverified.
		return rocco.NoBody{}, nil
	}

	// Generate magic link token.
	rawToken, err := intsession.GenerateToken()
	if err != nil {
		return rocco.NoBody{}, nil
	}
	now := time.Now()
	vt := &models.VerificationToken{
		Token:     rawToken,
		UserID:    user.ID,
		Type:      models.TokenTypeMagicLink,
		CreatedAt: now,
		ExpiresAt: now.Add(tokensCfg.MagicLinkTTL),
	}
	if err := verificationTokens.Set(req.Context, vt, tokensCfg.MagicLinkTTL); err != nil {
		return rocco.NoBody{}, nil
	}

	// Send magic link email (best-effort).
	emailClient := extpostmark.NewClient(postmarkCfg.ServerToken, postmarkCfg.DefaultFrom)
	_, _ = emailClient.SendEmail(req.Context, extpostmark.EmailRequest{
		To:       user.Email,
		Subject:  "Your sign-in link",
		TextBody: "Use the following token to sign in:\n\n" + rawToken + "\n\nThis token expires in 15 minutes.",
	})

	return rocco.NoBody{}, nil
}).WithSummary("Request magic link").
	WithDescription("Sends a magic link sign-in email. Always returns 204 regardless of whether the email exists.").
	WithTags("Auth").
	WithSuccessStatus(204)

// MagicLinkCallback validates a magic link token, creates a session, and redirects.
var MagicLinkCallback = rocco.GET("/login/magic/callback", func(req *rocco.Request[rocco.NoBody]) (rocco.Redirect, error) {
	verificationTokens := sum.MustUse[contracts.VerificationTokens](req.Context)
	sessions := sum.MustUse[contracts.Sessions](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	rawToken := req.Params.Query["token"]
	if rawToken == "" {
		return rocco.Redirect{}, ErrInvalidToken
	}

	// Validate the token.
	vt, err := verificationTokens.Get(req.Context, rawToken)
	if err != nil || vt == nil {
		return rocco.Redirect{}, ErrInvalidToken
	}
	if vt.Type != models.TokenTypeMagicLink || vt.IsExpired() {
		return rocco.Redirect{}, ErrInvalidToken
	}

	// Consume the token (single-use).
	_ = verificationTokens.Delete(req.Context, rawToken)

	// Create session.
	sessionToken, err := intsession.GenerateToken()
	if err != nil {
		return rocco.Redirect{}, ErrLoginFailed
	}
	now := time.Now()
	sess := &models.Session{
		Token:     sessionToken,
		UserID:    vt.UserID,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionCfg.TTL),
	}
	if err := sessions.SetWithUserIndex(req.Context, sess, sessionCfg.TTL); err != nil {
		return rocco.Redirect{}, ErrLoginFailed
	}

	headers := http.Header{}
	headers.Add("Set-Cookie", buildSessionCookie(sessionCfg, sessionToken).String())

	return rocco.Redirect{
		URL:     "/",
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("Magic link callback").
	WithDescription("Validates a magic link token, creates a session, and redirects with session cookie.").
	WithTags("Auth").
	WithQueryParams("token").
	WithErrors(ErrInvalidToken, ErrLoginFailed)

// VerifyEmail verifies a user's email address using a token.
// On success the user is logged in and redirected with a session cookie.
var VerifyEmail = rocco.POST("/verify-email", func(req *rocco.Request[wire.VerifyEmailRequest]) (rocco.Redirect, error) {
	users := sum.MustUse[contracts.Users](req.Context)
	verificationTokens := sum.MustUse[contracts.VerificationTokens](req.Context)
	sessions := sum.MustUse[contracts.Sessions](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	// Validate the token.
	vt, err := verificationTokens.Get(req.Context, req.Body.Token)
	if err != nil || vt == nil {
		return rocco.Redirect{}, ErrInvalidToken
	}
	if vt.Type != models.TokenTypeEmailVerify || vt.IsExpired() {
		return rocco.Redirect{}, ErrInvalidToken
	}

	// Consume the token (single-use).
	_ = verificationTokens.Delete(req.Context, req.Body.Token)

	// Mark email as verified.
	user, err := users.Get(req.Context, vt.UserID)
	if err != nil || user == nil {
		return rocco.Redirect{}, ErrUserNotFound
	}
	user.EmailVerified = true
	if err := users.Set(req.Context, user.ID, user); err != nil {
		return rocco.Redirect{}, ErrLoginFailed
	}

	// Create session so the user is immediately logged in.
	sessionToken, err := intsession.GenerateToken()
	if err != nil {
		return rocco.Redirect{}, ErrLoginFailed
	}
	now := time.Now()
	sess := &models.Session{
		Token:     sessionToken,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionCfg.TTL),
	}
	if err := sessions.SetWithUserIndex(req.Context, sess, sessionCfg.TTL); err != nil {
		return rocco.Redirect{}, ErrLoginFailed
	}

	headers := http.Header{}
	headers.Add("Set-Cookie", buildSessionCookie(sessionCfg, sessionToken).String())

	return rocco.Redirect{
		URL:     "/",
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("Verify email").
	WithDescription("Verifies a user's email address. Creates a session and redirects with session cookie on success.").
	WithTags("Auth").
	WithErrors(ErrInvalidToken, ErrUserNotFound, ErrLoginFailed)

// RequestPasswordReset sends a password reset email.
// Always responds 204 so callers cannot enumerate registered emails.
var RequestPasswordReset = rocco.POST("/password/reset", func(req *rocco.Request[wire.PasswordResetRequest]) (rocco.NoBody, error) {
	users := sum.MustUse[contracts.Users](req.Context)
	verificationTokens := sum.MustUse[contracts.VerificationTokens](req.Context)
	tokensCfg := sum.MustUse[config.Tokens](req.Context)
	postmarkCfg := sum.MustUse[config.Postmark](req.Context)

	user, err := users.GetByEmail(req.Context, req.Body.Email)
	if err != nil || user == nil {
		// Do not reveal whether the email exists.
		return rocco.NoBody{}, nil
	}

	// Generate reset token.
	rawToken, err := intsession.GenerateToken()
	if err != nil {
		return rocco.NoBody{}, nil
	}
	now := time.Now()
	vt := &models.VerificationToken{
		Token:     rawToken,
		UserID:    user.ID,
		Type:      models.TokenTypePasswordReset,
		CreatedAt: now,
		ExpiresAt: now.Add(tokensCfg.PasswordResetTTL),
	}
	if err := verificationTokens.Set(req.Context, vt, tokensCfg.PasswordResetTTL); err != nil {
		return rocco.NoBody{}, nil
	}

	// Send reset email (best-effort).
	emailClient := extpostmark.NewClient(postmarkCfg.ServerToken, postmarkCfg.DefaultFrom)
	_, _ = emailClient.SendEmail(req.Context, extpostmark.EmailRequest{
		To:       user.Email,
		Subject:  "Reset your password",
		TextBody: "Use the following token to reset your password:\n\n" + rawToken + "\n\nThis token expires in 1 hour.",
	})

	return rocco.NoBody{}, nil
}).WithSummary("Request password reset").
	WithDescription("Sends a password reset email. Always returns 204 regardless of whether the email exists.").
	WithTags("Auth").
	WithSuccessStatus(204)

// ConfirmPasswordReset completes a password reset using a token.
var ConfirmPasswordReset = rocco.POST("/password/reset/confirm", func(req *rocco.Request[wire.PasswordResetConfirmRequest]) (rocco.NoBody, error) {
	users := sum.MustUse[contracts.Users](req.Context)
	verificationTokens := sum.MustUse[contracts.VerificationTokens](req.Context)

	// Validate the token.
	vt, err := verificationTokens.Get(req.Context, req.Body.Token)
	if err != nil || vt == nil {
		return rocco.NoBody{}, ErrInvalidToken
	}
	if vt.Type != models.TokenTypePasswordReset || vt.IsExpired() {
		return rocco.NoBody{}, ErrInvalidToken
	}

	// Consume the token (single-use).
	_ = verificationTokens.Delete(req.Context, req.Body.Token)

	// Hash the new password.
	hash, err := intpassword.Hash(req.Body.Password)
	if err != nil {
		return rocco.NoBody{}, ErrLoginFailed
	}

	// Update the user's password.
	user, err := users.Get(req.Context, vt.UserID)
	if err != nil || user == nil {
		return rocco.NoBody{}, ErrUserNotFound
	}
	user.PasswordHash = &hash
	if err := users.Set(req.Context, user.ID, user); err != nil {
		return rocco.NoBody{}, ErrLoginFailed
	}

	return rocco.NoBody{}, nil
}).WithSummary("Confirm password reset").
	WithDescription("Completes a password reset. The user may now log in with the new password.").
	WithTags("Auth").
	WithErrors(ErrInvalidToken, ErrUserNotFound, ErrLoginFailed)
