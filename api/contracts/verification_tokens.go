package contracts

import (
	"context"
	"time"

	"github.com/zoobzio/sumatra/models"
)

// VerificationTokens defines the contract for verification token operations
// required by the public API.
type VerificationTokens interface {
	// Get retrieves a verification token by its token string.
	Get(ctx context.Context, token string) (*models.VerificationToken, error)
	// Set stores a verification token with the given TTL.
	Set(ctx context.Context, token *models.VerificationToken, ttl time.Duration) error
	// Delete removes a verification token by its token string.
	Delete(ctx context.Context, token string) error
}
