package contracts

import "context"

// Providers defines the contract for OAuth provider operations required by the admin API.
type Providers interface {
	// DeleteByUser removes all provider links for the given userID.
	DeleteByUser(ctx context.Context, userID string) error
}
