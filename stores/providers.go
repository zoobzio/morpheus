package stores

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zoobzio/astql"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/models"
)

// Providers provides database access for OAuth provider link records.
type Providers struct {
	*sum.Database[models.Provider]
}

// NewProviders creates a new providers store backed by PostgreSQL.
func NewProviders(db *sqlx.DB, renderer astql.Renderer) (*Providers, error) {
	database, err := sum.NewDatabase[models.Provider](db, "providers", renderer)
	if err != nil {
		return nil, err
	}
	return &Providers{Database: database}, nil
}

// GetByUserAndType retrieves the provider link for a given user and provider type.
func (s *Providers) GetByUserAndType(ctx context.Context, userID string, providerType models.ProviderType) (*models.Provider, error) {
	return s.Select().
		Where("user_id", "=", "user_id").
		Where("type", "=", "type").
		Exec(ctx, map[string]any{
			"user_id": userID,
			"type":    string(providerType),
		})
}

// GetByProviderUser retrieves a provider link by external provider type and user ID.
// Used during OAuth callbacks to match incoming tokens to existing accounts.
func (s *Providers) GetByProviderUser(ctx context.Context, providerType models.ProviderType, providerUserID string) (*models.Provider, error) {
	return s.Select().
		Where("type", "=", "type").
		Where("provider_user_id", "=", "provider_user_id").
		Exec(ctx, map[string]any{
			"type":             string(providerType),
			"provider_user_id": providerUserID,
		})
}

// ListByUser retrieves all provider links for a given user.
func (s *Providers) ListByUser(ctx context.Context, userID string) ([]*models.Provider, error) {
	return s.Query().
		Where("user_id", "=", "user_id").
		Exec(ctx, map[string]any{"user_id": userID})
}

// DeleteByUserAndType removes the provider link for a specific user and provider type.
func (s *Providers) DeleteByUserAndType(ctx context.Context, userID string, providerType models.ProviderType) error {
	_, err := s.Remove().
		Where("user_id", "=", "user_id").
		Where("type", "=", "type").
		Exec(ctx, map[string]any{
			"user_id": userID,
			"type":    string(providerType),
		})
	return err
}

// DeleteByUser removes all provider links for a given user.
func (s *Providers) DeleteByUser(ctx context.Context, userID string) error {
	_, err := s.Remove().
		Where("user_id", "=", "user_id").
		Exec(ctx, map[string]any{"user_id": userID})
	return err
}
