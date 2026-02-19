package contracts

import (
	"context"

	"github.com/zoobzio/sumatra/models"
)

// Sessions defines the contract for session operations required by the admin API.
type Sessions interface {
	// Get retrieves a session by its token.
	Get(ctx context.Context, token string) (*models.Session, error)
	// Delete removes a session by its token.
	Delete(ctx context.Context, token string) error
	// ListByUser returns up to limit session tokens belonging to userID.
	ListByUser(ctx context.Context, userID string, limit int) ([]string, error)
	// DeleteByUser revokes all sessions for the given userID.
	DeleteByUser(ctx context.Context, userID string) error
}
