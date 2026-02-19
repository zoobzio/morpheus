// Package transformers provides pure functions for mapping between domain models
// and public API wire types.
package transformers

import (
	"github.com/zoobzio/sumatra/api/wire"
	"github.com/zoobzio/sumatra/models"
)

// UserToResponse transforms a User model to a public API UserResponse.
func UserToResponse(u *models.User) wire.UserResponse {
	return wire.UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Name:          u.Name,
		AvatarURL:     u.AvatarURL,
	}
}

// ApplyUserUpdate applies the fields from a UserUpdateRequest onto an existing User model.
// Only non-nil fields are applied.
func ApplyUserUpdate(req wire.UserUpdateRequest, u *models.User) {
	if req.Name != nil {
		u.Name = req.Name
	}
}
