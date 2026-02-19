// Package contracts defines the interfaces consumed by the public API handlers.
package contracts

import (
	"context"

	"github.com/zoobzio/sumatra/models"
)

// Users defines the contract for user operations required by the public API.
type Users interface {
	// Get retrieves a user by primary key.
	Get(ctx context.Context, key string) (*models.User, error)
	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	// Set creates or updates a user record.
	Set(ctx context.Context, key string, user *models.User) error
}
