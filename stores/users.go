package stores

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zoobzio/astql"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/models"
)

// Users provides database access for user records.
type Users struct {
	*sum.Database[models.User]
}

// NewUsers creates a new users store backed by PostgreSQL.
func NewUsers(db *sqlx.DB, renderer astql.Renderer) (*Users, error) {
	database, err := sum.NewDatabase[models.User](db, "users", renderer)
	if err != nil {
		return nil, err
	}
	return &Users{Database: database}, nil
}

// GetByEmail retrieves a user by their email address.
func (s *Users) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.Select().
		Where("email", "=", "email").
		Exec(ctx, map[string]any{"email": email})
}

// List returns a paginated list of users ordered by created_at DESC.
func (s *Users) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return s.Query().
		OrderBy("created_at", "DESC").
		Limit(limit).
		Offset(offset).
		Exec(ctx, nil)
}

// Count returns the total number of users.
func (s *Users) Count(ctx context.Context) (float64, error) {
	return s.Database.Count().Exec(ctx, nil)
}
