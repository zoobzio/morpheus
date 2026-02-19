package handlers

import (
	"strconv"

	"github.com/zoobzio/rocco"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/admin/contracts"
	"github.com/zoobzio/sumatra/admin/transformers"
	"github.com/zoobzio/sumatra/admin/wire"
)

// ListUsers returns a paginated list of all users in the system.
// Accepts optional query parameters: limit (default 50) and offset (default 0).
var ListUsers = rocco.GET("/users", func(req *rocco.Request[rocco.NoBody]) (wire.AdminUserListResponse, error) {
	users := sum.MustUse[contracts.Users](req.Context)

	limit := 50
	if l := req.Params.Query["limit"]; l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := req.Params.Query["offset"]; o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	list, err := users.List(req.Context, limit, offset)
	if err != nil {
		return wire.AdminUserListResponse{}, err
	}

	total, err := users.Count(req.Context)
	if err != nil {
		return wire.AdminUserListResponse{}, err
	}

	return transformers.UsersToAdminList(list, int(total)), nil
}).WithSummary("List users").
	WithDescription("Returns a paginated list of all users in the system.").
	WithTags("Users").
	WithQueryParams("limit", "offset").
	WithAuthentication()

// GetUser returns a single user by ID.
var GetUser = rocco.GET("/users/{id}", func(req *rocco.Request[rocco.NoBody]) (wire.AdminUserResponse, error) {
	users := sum.MustUse[contracts.Users](req.Context)

	id := req.Params.Path["id"]

	user, err := users.Get(req.Context, id)
	if err != nil {
		return wire.AdminUserResponse{}, ErrUserNotFound
	}

	return transformers.UserToAdminResponse(user), nil
}).WithSummary("Get user").
	WithDescription("Returns a single user by ID.").
	WithTags("Users").
	WithPathParams("id").
	WithErrors(ErrUserNotFound).
	WithAuthentication()

// DeleteUser removes a user and cascades to their sessions and provider links.
var DeleteUser = rocco.DELETE("/users/{id}", func(req *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	users := sum.MustUse[contracts.Users](req.Context)
	sessions := sum.MustUse[contracts.Sessions](req.Context)
	providers := sum.MustUse[contracts.Providers](req.Context)

	id := req.Params.Path["id"]

	// Verify the user exists before cascading.
	if _, err := users.Get(req.Context, id); err != nil {
		return rocco.NoBody{}, ErrUserNotFound
	}

	// Cascade: revoke all sessions.
	if err := sessions.DeleteByUser(req.Context, id); err != nil {
		return rocco.NoBody{}, err
	}

	// Cascade: remove all provider links.
	if err := providers.DeleteByUser(req.Context, id); err != nil {
		return rocco.NoBody{}, err
	}

	// Delete the user record.
	if err := users.Delete(req.Context, id); err != nil {
		return rocco.NoBody{}, err
	}

	return rocco.NoBody{}, nil
}).WithSummary("Delete user").
	WithDescription("Deletes a user and cascades the deletion to their sessions and OAuth provider links.").
	WithTags("Users").
	WithPathParams("id").
	WithErrors(ErrUserNotFound).
	WithAuthentication().
	WithSuccessStatus(204)
