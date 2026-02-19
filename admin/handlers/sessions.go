package handlers

import (
	"github.com/zoobzio/rocco"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/admin/contracts"
	"github.com/zoobzio/sumatra/admin/transformers"
	"github.com/zoobzio/sumatra/admin/wire"
	"github.com/zoobzio/sumatra/models"
)

// ListSessions returns all sessions for a given user_id query parameter.
var ListSessions = rocco.GET("/sessions", func(req *rocco.Request[rocco.NoBody]) (wire.AdminSessionListResponse, error) {
	sessions := sum.MustUse[contracts.Sessions](req.Context)

	userID := req.Params.Query["user_id"]
	if userID == "" {
		return wire.AdminSessionListResponse{}, rocco.ErrBadRequest.WithMessage("query parameter 'user_id' is required")
	}

	tokens, err := sessions.ListByUser(req.Context, userID, 0)
	if err != nil {
		return wire.AdminSessionListResponse{}, err
	}

	records := make([]*models.Session, 0, len(tokens))
	for _, token := range tokens {
		s, err := sessions.Get(req.Context, token)
		if err != nil || s == nil {
			continue
		}
		records = append(records, s)
	}

	return transformers.SessionsToAdminList(records), nil
}).WithSummary("List sessions").
	WithDescription("Returns all active sessions for the specified user. The user_id query parameter is required.").
	WithTags("Sessions").
	WithQueryParams("user_id").
	WithAuthentication()

// RevokeSession deletes a specific session by its token.
var RevokeSession = rocco.DELETE("/sessions/{token}", func(req *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
	sessions := sum.MustUse[contracts.Sessions](req.Context)

	token := req.Params.Path["token"]

	if _, err := sessions.Get(req.Context, token); err != nil {
		return rocco.NoBody{}, ErrSessionNotFound
	}

	if err := sessions.Delete(req.Context, token); err != nil {
		return rocco.NoBody{}, err
	}

	return rocco.NoBody{}, nil
}).WithSummary("Revoke session").
	WithDescription("Revokes a specific session by its token.").
	WithTags("Sessions").
	WithPathParams("token").
	WithErrors(ErrSessionNotFound).
	WithAuthentication().
	WithSuccessStatus(204)
