// Package wire contains the admin API request and response types.
package wire

import "time"

// AdminUserResponse is the admin API response for user data.
// All fields are fully visible — no masking for internal staff.
type AdminUserResponse struct {
	ID            string    `json:"id" description:"User UUID" example:"01942d3a-1234-7abc-8def-0123456789ab"`
	Email         string    `json:"email" description:"Email address (unmasked)" example:"user@example.com"`
	EmailVerified bool      `json:"email_verified" description:"Whether the email address has been verified"`
	Name          *string   `json:"name,omitempty" description:"Display name" example:"Jane Doe"`
	AvatarURL     *string   `json:"avatar_url,omitempty" description:"Avatar URL" example:"https://avatars.githubusercontent.com/u/1"`
	CreatedAt     time.Time `json:"created_at" description:"Account creation time"`
	UpdatedAt     time.Time `json:"updated_at" description:"Last update time"`
}

// Clone returns a deep copy of AdminUserResponse.
func (u AdminUserResponse) Clone() AdminUserResponse {
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

// AdminUserListResponse is the admin API response for a paginated list of users.
type AdminUserListResponse struct {
	Users []AdminUserResponse `json:"users" description:"List of user records"`
	Total int                 `json:"total" description:"Total number of users in the system" example:"1024"`
}

// Clone returns a deep copy of AdminUserListResponse.
func (r AdminUserListResponse) Clone() AdminUserListResponse {
	c := r
	if r.Users != nil {
		c.Users = make([]AdminUserResponse, len(r.Users))
		for i, u := range r.Users {
			c.Users[i] = u.Clone()
		}
	}
	return c
}
