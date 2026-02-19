package models

import (
	"context"
	"time"

	"github.com/zoobzio/check"
	"github.com/zoobzio/sum"
)

// ProviderType represents the type of OAuth provider.
type ProviderType string

const (
	// ProviderTypeGitHub is the GitHub OAuth provider.
	ProviderTypeGitHub ProviderType = "github"
	// ProviderTypeGoogle is the Google OAuth provider.
	ProviderTypeGoogle ProviderType = "google"
)

// Provider represents an OAuth provider linked to a user account.
type Provider struct {
	ID             int64        `json:"id" db:"id" constraints:"primarykey" description:"Auto-increment primary key" example:"1"`
	UserID         string       `json:"user_id" db:"user_id" constraints:"notnull" references:"users(id)" description:"FK to users.id" example:"01942d3a-1234-7abc-8def-0123456789ab"`
	Type           ProviderType `json:"type" db:"type" constraints:"notnull" description:"OAuth provider type" example:"github"`
	ProviderUserID string       `json:"provider_user_id" db:"provider_user_id" constraints:"notnull" description:"External provider user ID" example:"12345678"`
	AccessToken    string       `json:"-" db:"access_token" constraints:"notnull" store.encrypt:"aes" load.decrypt:"aes" description:"Encrypted OAuth access token"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at" constraints:"notnull" default:"now()" description:"Record creation time"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at" constraints:"notnull" default:"now()" description:"Last update time"`
}

// BeforeSave encrypts sensitive fields before database write.
func (p *Provider) BeforeSave(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[Provider]](ctx)
	stored, err := b.Store(ctx, *p)
	if err != nil {
		return err
	}
	*p = stored
	return nil
}

// AfterLoad decrypts sensitive fields after database read.
func (p *Provider) AfterLoad(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[Provider]](ctx)
	loaded, err := b.Load(ctx, *p)
	if err != nil {
		return err
	}
	*p = loaded
	return nil
}

// Validate validates the Provider model.
func (p Provider) Validate() error {
	return check.All(
		check.Str(p.UserID, "user_id").Required().V(),
		check.Str(string(p.Type), "type").Required().OneOf([]string{"github", "google"}).V(),
		check.Str(p.ProviderUserID, "provider_user_id").Required().V(),
		check.Str(p.AccessToken, "access_token").Required().V(),
	).Err()
}

// Clone returns a deep copy of the Provider.
func (p Provider) Clone() Provider {
	return p
}
