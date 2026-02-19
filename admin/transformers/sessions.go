package transformers

import (
	"github.com/zoobzio/sumatra/admin/wire"
	"github.com/zoobzio/sumatra/models"
)

// maskToken reduces a session token to its first 8 characters followed by "..."
// so that it can be displayed in the admin UI without exposing the full secret.
func maskToken(token string) string {
	if len(token) <= 8 {
		return token + "..."
	}
	return token[:8] + "..."
}

// SessionToAdminResponse transforms a Session model to an AdminSessionResponse.
// The token is partially masked for display — first 8 characters + "...".
func SessionToAdminResponse(s *models.Session) wire.AdminSessionResponse {
	return wire.AdminSessionResponse{
		Token:     maskToken(s.Token),
		UserID:    s.UserID,
		CreatedAt: s.CreatedAt,
		ExpiresAt: s.ExpiresAt,
	}
}

// SessionsToAdminList transforms a slice of Session models to an
// AdminSessionListResponse. Each token is partially masked.
func SessionsToAdminList(sessions []*models.Session) wire.AdminSessionListResponse {
	resp := wire.AdminSessionListResponse{
		Sessions: make([]wire.AdminSessionResponse, len(sessions)),
	}
	for i, s := range sessions {
		resp.Sessions[i] = SessionToAdminResponse(s)
	}
	return resp
}
