package handlers

import (
	"net/http"

	"github.com/zoobzio/rocco"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/api/contracts"
	"github.com/zoobzio/sumatra/api/transformers"
	"github.com/zoobzio/sumatra/api/wire"
	"github.com/zoobzio/sumatra/config"
)

// GetMe returns the authenticated user's profile.
var GetMe = rocco.GET("/me", func(req *rocco.Request[rocco.NoBody]) (wire.UserResponse, error) {
	users := sum.MustUse[contracts.Users](req.Context)

	user, err := users.Get(req.Context, req.Identity.ID())
	if err != nil {
		return wire.UserResponse{}, ErrUserNotFound
	}

	return transformers.UserToResponse(user), nil
}).WithSummary("Get current user").
	WithDescription("Returns the authenticated user's profile.").
	WithTags("Users").
	WithAuthentication().
	WithErrors(ErrUserNotFound)

// UpdateMe updates the authenticated user's display name.
var UpdateMe = rocco.PATCH("/me", func(req *rocco.Request[wire.UserUpdateRequest]) (wire.UserResponse, error) {
	users := sum.MustUse[contracts.Users](req.Context)

	user, err := users.Get(req.Context, req.Identity.ID())
	if err != nil {
		return wire.UserResponse{}, ErrUserNotFound
	}

	transformers.ApplyUserUpdate(req.Body, user)

	if err := users.Set(req.Context, user.ID, user); err != nil {
		return wire.UserResponse{}, err
	}

	return transformers.UserToResponse(user), nil
}).WithSummary("Update current user").
	WithDescription("Updates the authenticated user's display name.").
	WithTags("Users").
	WithAuthentication().
	WithErrors(ErrUserNotFound)

// Logout invalidates the current session and redirects with a cleared cookie.
var Logout = rocco.POST("/logout", func(req *rocco.Request[rocco.NoBody]) (rocco.Redirect, error) {
	sessions := sum.MustUse[contracts.Sessions](req.Context)
	sessionCfg := sum.MustUse[config.Session](req.Context)

	// Retrieve token from cookie.
	cookie, err := req.Cookie(sessionCfg.CookieName)
	if err == nil && cookie != nil {
		// Best-effort delete — don't fail logout if session already gone.
		_ = sessions.Delete(req.Context, cookie.Value)
	}

	// Clear the session cookie.
	clearCookie := &http.Cookie{
		Name:     sessionCfg.CookieName,
		Value:    "",
		Path:     sessionCfg.CookiePath,
		Domain:   sessionCfg.CookieDomain,
		MaxAge:   -1,
		Secure:   sessionCfg.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	headers := http.Header{}
	headers.Add("Set-Cookie", clearCookie.String())

	return rocco.Redirect{
		URL:     "/",
		Status:  http.StatusFound,
		Headers: headers,
	}, nil
}).WithSummary("Logout").
	WithDescription("Invalidates the current session and redirects with cleared cookie.").
	WithTags("Auth")
