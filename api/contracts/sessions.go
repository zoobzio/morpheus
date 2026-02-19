package contracts

import (
	"context"
	"time"

	"github.com/zoobzio/sumatra/models"
)

// Sessions defines the contract for session operations required by the public API.
type Sessions interface {
	// Get retrieves a session by its token.
	Get(ctx context.Context, token string) (*models.Session, error)
	// SetWithUserIndex stores a session and writes a corresponding user index entry.
	// The user index enables future enumeration and bulk-revocation of sessions.
	SetWithUserIndex(ctx context.Context, session *models.Session, ttl time.Duration) error
	// Delete removes a session by its token.
	Delete(ctx context.Context, token string) error
}
