package handlers

import "github.com/zoobzio/rocco"

// All returns all public API endpoints for registration with the router.
func All() []rocco.Endpoint {
	return []rocco.Endpoint{
		// Auth
		Register,
		Login,
		RequestMagicLink,
		MagicLinkCallback,
		VerifyEmail,
		RequestPasswordReset,
		ConfirmPasswordReset,
		Logout,
		InitiateGitHubLogin,
		GitHubLoginCallback,
		InitiateGoogleLogin,
		GoogleLoginCallback,

		// Users
		GetMe,
		UpdateMe,

		// Providers
		ListProviders,
		InitiateGitHubLink,
		GitHubLinkCallback,
		UnlinkGitHub,
		InitiateGoogleLink,
		GoogleLinkCallback,
		UnlinkGoogle,
	}
}
