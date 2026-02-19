package identity

import (
	"github.com/zoobzio/aegis/proto/identity"
	"github.com/zoobzio/sumatra/models"
)

func userToProto(u *models.User) *identity.GetUserResponse {
	resp := &identity.GetUserResponse{
		Id:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt.Unix(),
	}
	if u.Name != nil {
		resp.Name = *u.Name
	}
	if u.AvatarURL != nil {
		resp.AvatarUrl = *u.AvatarURL
	}
	return resp
}
