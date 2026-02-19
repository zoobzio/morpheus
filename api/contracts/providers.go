package contracts

import (
	"context"

	"github.com/zoobzio/sumatra/models"
)

// Providers defines the contract for OAuth provider operations required by the public API.
type Providers interface {
	// GetByUserAndType retrieves the provider link for a given user and provider type.
	GetByUserAndType(ctx context.Context, userID string, providerType models.ProviderType) (*models.Provider, error)
	// GetByProviderUser retrieves a provider link by external provider type and user ID.
	// Used during OAuth callbacks to match incoming tokens to existing accounts.
	GetByProviderUser(ctx context.Context, providerType models.ProviderType, providerUserID string) (*models.Provider, error)
	// Set creates or updates a provider link record.
	Set(ctx context.Context, key string, provider *models.Provider) error
	// DeleteByUserAndType removes the provider link for a specific user and provider type.
	DeleteByUserAndType(ctx context.Context, userID string, providerType models.ProviderType) error
	// ListByUser retrieves all provider links for a given user.
	ListByUser(ctx context.Context, userID string) ([]*models.Provider, error)
}
