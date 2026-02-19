// Package transformers provides pure functions for mapping between domain models
// and admin API wire types.
package transformers

import (
	"github.com/zoobzio/sumatra/admin/wire"
	"github.com/zoobzio/sumatra/models"
)

// UserToAdminResponse transforms a User model to an AdminUserResponse.
// All fields are mapped without masking — the admin surface has full visibility.
func UserToAdminResponse(u *models.User) wire.AdminUserResponse {
	return wire.AdminUserResponse{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Name:          u.Name,
		AvatarURL:     u.AvatarURL,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// UsersToAdminList transforms a slice of User models and a total count to an
// AdminUserListResponse.
func UsersToAdminList(users []*models.User, total int) wire.AdminUserListResponse {
	resp := wire.AdminUserListResponse{
		Users: make([]wire.AdminUserResponse, len(users)),
		Total: total,
	}
	for i, u := range users {
		resp.Users[i] = UserToAdminResponse(u)
	}
	return resp
}
