// Package contracts defines the interfaces consumed by the admin API handlers.
package contracts

import (
	"context"

	"github.com/zoobzio/sumatra/models"
)

// Users defines the contract for user operations required by the admin API.
type Users interface {
	// Get retrieves a user by primary key.
	Get(ctx context.Context, key string) (*models.User, error)
	// List returns a paginated list of users ordered by created_at DESC.
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	// Count returns the total number of users.
	Count(ctx context.Context) (float64, error)
	// Delete removes a user by primary key.
	Delete(ctx context.Context, key string) error
}
