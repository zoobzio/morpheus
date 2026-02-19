// Package handlers contains the admin API HTTP handlers.
package handlers

import "github.com/zoobzio/rocco"

var (
	// ErrUserNotFound is returned when a requested user does not exist.
	ErrUserNotFound = rocco.ErrNotFound.WithMessage("user not found")
	// ErrSessionNotFound is returned when a requested session token cannot be found.
	ErrSessionNotFound = rocco.ErrNotFound.WithMessage("session not found")
)
