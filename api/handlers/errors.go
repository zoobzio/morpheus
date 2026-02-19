// Package handlers contains the public API HTTP handlers.
package handlers

import "github.com/zoobzio/rocco"

var (
	// ErrUserNotFound is returned when a requested user does not exist.
	ErrUserNotFound = rocco.ErrNotFound.WithMessage("user not found")
	// ErrSessionNotFound is returned when a session token cannot be found.
	ErrSessionNotFound = rocco.ErrNotFound.WithMessage("session not found")
	// ErrSessionExpired is returned when the session token exists but has expired.
	ErrSessionExpired = rocco.ErrUnauthorized.WithMessage("session expired")
	// ErrInvalidCredentials is returned when email/password do not match.
	ErrInvalidCredentials = rocco.ErrUnauthorized.WithMessage("invalid email or password")
	// ErrEmailNotVerified is returned when a login is attempted before verifying email.
	ErrEmailNotVerified = rocco.ErrForbidden.WithMessage("email address not verified")
	// ErrInvalidToken is returned when a verification token is missing, expired, or has wrong type.
	ErrInvalidToken = rocco.ErrBadRequest.WithMessage("invalid or expired token")
	// ErrEmailAlreadyExists is returned when registering with an email that is already in use.
	ErrEmailAlreadyExists = rocco.ErrConflict.WithMessage("email address already registered")
	// ErrRegistrationFailed is returned when user creation fails for an unexpected reason.
	ErrRegistrationFailed = rocco.ErrInternalServer.WithMessage("registration failed")
	// ErrLoginFailed is returned when session creation fails for an unexpected reason.
	ErrLoginFailed = rocco.ErrInternalServer.WithMessage("login failed")

	// ErrProviderAlreadyLinked is returned when a provider is already linked to a different account.
	ErrProviderAlreadyLinked = rocco.ErrConflict.WithMessage("provider already linked to another account")
	// ErrProviderNotFound is returned when a provider link does not exist.
	ErrProviderNotFound = rocco.ErrNotFound.WithMessage("provider not found")
	// ErrProviderLinkFailed is returned when linking a provider fails for an unexpected reason.
	ErrProviderLinkFailed = rocco.ErrInternalServer.WithMessage("failed to link provider")
	// ErrLastAuthMethod is returned when the user tries to unlink their only authentication method.
	ErrLastAuthMethod = rocco.ErrConflict.WithMessage("cannot unlink last authentication method")
	// ErrGitHubOAuthFailed is returned when the GitHub OAuth exchange or user fetch fails.
	ErrGitHubOAuthFailed = rocco.ErrInternalServer.WithMessage("github oauth failed")
	// ErrGoogleOAuthFailed is returned when the Google OAuth exchange or user fetch fails.
	ErrGoogleOAuthFailed = rocco.ErrInternalServer.WithMessage("google oauth failed")
	// ErrAccountNotLinked is returned on provider login when no account is linked to that identity.
	ErrAccountNotLinked = rocco.ErrUnauthorized.WithMessage("no account linked to this provider account")
)
