// Package wire contains the public API request and response types.
package wire

import (
	"context"

	"github.com/zoobzio/check"
	"github.com/zoobzio/sum"
)

// UserResponse is the public API response for user data.
// Email is masked before the response is sent to the client.
type UserResponse struct {
	ID            string  `json:"id" description:"User UUID" example:"01942d3a-1234-7abc-8def-0123456789ab"`
	Email         string  `json:"email" description:"Email address (masked)" example:"u***@example.com" send.mask:"email"`
	EmailVerified bool    `json:"email_verified" description:"Whether the email address has been verified"`
	Name          *string `json:"name,omitempty" description:"Display name" example:"Jane Doe"`
	AvatarURL     *string `json:"avatar_url,omitempty" description:"Avatar URL" example:"https://avatars.githubusercontent.com/u/1"`
}

// OnSend applies boundary masking before the response is marshaled.
// Implements rocco.Sendable.
func (u *UserResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[UserResponse]](ctx)
	masked, err := b.Send(ctx, *u)
	if err != nil {
		return err
	}
	*u = masked
	return nil
}

// Clone returns a deep copy of UserResponse.
func (u UserResponse) Clone() UserResponse {
	c := u
	if u.Name != nil {
		n := *u.Name
		c.Name = &n
	}
	if u.AvatarURL != nil {
		a := *u.AvatarURL
		c.AvatarURL = &a
	}
	return c
}

// UserUpdateRequest is the request body for updating a user's profile.
type UserUpdateRequest struct {
	Name *string `json:"name,omitempty" description:"New display name" example:"Jane Doe"`
}

// Validate validates the UserUpdateRequest.
func (r *UserUpdateRequest) Validate() error {
	return check.All(
		check.OptStr(r.Name, "name").MaxLen(255).V(),
	).Err()
}

// Clone returns a deep copy of UserUpdateRequest.
func (r UserUpdateRequest) Clone() UserUpdateRequest {
	c := r
	if r.Name != nil {
		n := *r.Name
		c.Name = &n
	}
	return c
}
