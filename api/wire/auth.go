package wire

import "github.com/zoobzio/check"

// RegisterRequest is the request body for creating a new account.
type RegisterRequest struct {
	Email    string `json:"email" description:"Email address" example:"user@example.com"`
	Password string `json:"password" description:"Password (min 8 characters)" example:"correct-horse-battery"`
}

// Validate validates the RegisterRequest.
func (r *RegisterRequest) Validate() error {
	return check.All(
		check.Str(r.Email, "email").Required().Email().V(),
		check.Str(r.Password, "password").Required().MinLen(8).V(),
	).Err()
}

// Clone returns a deep copy of RegisterRequest.
func (r RegisterRequest) Clone() RegisterRequest {
	return r
}

// LoginRequest is the request body for password-based login.
type LoginRequest struct {
	Email    string `json:"email" description:"Email address" example:"user@example.com"`
	Password string `json:"password" description:"Password" example:"correct-horse-battery"`
}

// Validate validates the LoginRequest.
func (r *LoginRequest) Validate() error {
	return check.All(
		check.Str(r.Email, "email").Required().Email().V(),
		check.Str(r.Password, "password").Required().V(),
	).Err()
}

// Clone returns a deep copy of LoginRequest.
func (r LoginRequest) Clone() LoginRequest {
	return r
}

// MagicLinkRequest is the request body for requesting a magic link.
type MagicLinkRequest struct {
	Email string `json:"email" description:"Email address" example:"user@example.com"`
}

// Validate validates the MagicLinkRequest.
func (r *MagicLinkRequest) Validate() error {
	return check.All(
		check.Str(r.Email, "email").Required().Email().V(),
	).Err()
}

// Clone returns a deep copy of MagicLinkRequest.
func (r MagicLinkRequest) Clone() MagicLinkRequest {
	return r
}

// VerifyEmailRequest is the request body for verifying an email address.
type VerifyEmailRequest struct {
	Token string `json:"token" description:"Email verification token" example:"dGhpcyBpcyBhIHRva2Vu"`
}

// Validate validates the VerifyEmailRequest.
func (r *VerifyEmailRequest) Validate() error {
	return check.All(
		check.Str(r.Token, "token").Required().V(),
	).Err()
}

// Clone returns a deep copy of VerifyEmailRequest.
func (r VerifyEmailRequest) Clone() VerifyEmailRequest {
	return r
}

// PasswordResetRequest is the request body for requesting a password reset.
type PasswordResetRequest struct {
	Email string `json:"email" description:"Email address" example:"user@example.com"`
}

// Validate validates the PasswordResetRequest.
func (r *PasswordResetRequest) Validate() error {
	return check.All(
		check.Str(r.Email, "email").Required().Email().V(),
	).Err()
}

// Clone returns a deep copy of PasswordResetRequest.
func (r PasswordResetRequest) Clone() PasswordResetRequest {
	return r
}

// PasswordResetConfirmRequest is the request body for completing a password reset.
type PasswordResetConfirmRequest struct {
	Token    string `json:"token" description:"Password reset token" example:"dGhpcyBpcyBhIHRva2Vu"`
	Password string `json:"password" description:"New password (min 8 characters)" example:"correct-horse-battery"`
}

// Validate validates the PasswordResetConfirmRequest.
func (r *PasswordResetConfirmRequest) Validate() error {
	return check.All(
		check.Str(r.Token, "token").Required().V(),
		check.Str(r.Password, "password").Required().MinLen(8).V(),
	).Err()
}

// Clone returns a deep copy of PasswordResetConfirmRequest.
func (r PasswordResetConfirmRequest) Clone() PasswordResetConfirmRequest {
	return r
}
