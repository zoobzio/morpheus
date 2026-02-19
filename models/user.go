// Package models contains the domain models for morpheus.
package models

import (
	"time"

	"github.com/zoobzio/check"
)

// User represents an authenticated user account.
type User struct {
	ID            string    `json:"id" db:"id" constraints:"primarykey" description:"UUID v7 primary key" example:"01942d3a-1234-7abc-8def-0123456789ab"`
	Email         string    `json:"email" db:"email" constraints:"notnull,unique" description:"Email address" example:"user@example.com"`
	PasswordHash  *string   `json:"-" db:"password_hash" description:"argon2id hash, null for passwordless users"`
	EmailVerified bool      `json:"email_verified" db:"email_verified" constraints:"notnull" default:"false" description:"Whether the email address has been verified"`
	Name          *string   `json:"name,omitempty" db:"name" description:"Display name" example:"Jane Doe"`
	AvatarURL     *string   `json:"avatar_url,omitempty" db:"avatar_url" description:"Avatar URL" example:"https://avatars.githubusercontent.com/u/1"`
	CreatedAt     time.Time `json:"created_at" db:"created_at" constraints:"notnull" default:"now()" description:"Account creation time"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at" constraints:"notnull" default:"now()" description:"Last update time"`
}

// Validate validates the User model.
func (u User) Validate() error {
	return check.All(
		check.Str(u.ID, "id").Required().V(),
		check.Str(u.Email, "email").Required().V(),
	).Err()
}

// Clone returns a deep copy of the User.
func (u User) Clone() User {
	c := u
	if u.PasswordHash != nil {
		h := *u.PasswordHash
		c.PasswordHash = &h
	}
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
