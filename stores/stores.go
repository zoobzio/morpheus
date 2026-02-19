// Package stores provides data access implementations for morpheus.
// All stores are shared across API surfaces; individual contracts expose
// the subset of methods each surface requires.
package stores

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zoobzio/astql"
	"github.com/zoobzio/grub"
)

// Stores is the aggregate of all application data stores.
type Stores struct {
	Users              *Users
	Providers          *Providers
	Sessions           *Sessions
	VerificationTokens *VerificationTokens
}

// New initialises all stores and returns the aggregate.
// db and renderer are required for PostgreSQL-backed stores.
// sessionProvider is required for the Redis-backed sessions and verification token stores.
func New(db *sqlx.DB, renderer astql.Renderer, sessionProvider grub.StoreProvider) (*Stores, error) {
	users, err := NewUsers(db, renderer)
	if err != nil {
		return nil, fmt.Errorf("stores: failed to create users store: %w", err)
	}

	providers, err := NewProviders(db, renderer)
	if err != nil {
		return nil, fmt.Errorf("stores: failed to create providers store: %w", err)
	}

	sessions, err := NewSessions(sessionProvider)
	if err != nil {
		return nil, fmt.Errorf("stores: failed to create sessions store: %w", err)
	}

	verificationTokens, err := NewVerificationTokens(sessionProvider)
	if err != nil {
		return nil, fmt.Errorf("stores: failed to create verification tokens store: %w", err)
	}

	return &Stores{
		Users:              users,
		Providers:          providers,
		Sessions:           sessions,
		VerificationTokens: verificationTokens,
	}, nil
}
