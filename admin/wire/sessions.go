package wire

import "time"

// AdminSessionResponse is the admin API response for a session record.
// The token is partially masked (first 8 characters + "...") for display safety.
type AdminSessionResponse struct {
	Token     string    `json:"token" description:"Session token (partially masked)" example:"a1b2c3d4..."`
	UserID    string    `json:"user_id" description:"ID of the owning user" example:"01942d3a-1234-7abc-8def-0123456789ab"`
	CreatedAt time.Time `json:"created_at" description:"Session creation time"`
	ExpiresAt time.Time `json:"expires_at" description:"Session expiry time"`
}

// Clone returns a deep copy of AdminSessionResponse.
func (s AdminSessionResponse) Clone() AdminSessionResponse {
	return s
}

// AdminSessionListResponse is the admin API response for a list of sessions.
type AdminSessionListResponse struct {
	Sessions []AdminSessionResponse `json:"sessions" description:"List of session records"`
}

// Clone returns a deep copy of AdminSessionListResponse.
func (r AdminSessionListResponse) Clone() AdminSessionListResponse {
	c := r
	if r.Sessions != nil {
		c.Sessions = make([]AdminSessionResponse, len(r.Sessions))
		copy(c.Sessions, r.Sessions)
	}
	return c
}
