package stores

import (
	"context"
	"time"

	"github.com/zoobzio/grub"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/models"
)

const verificationPrefix = "verification:"

// verificationKey returns the Redis key for a verification token.
func verificationKey(token string) string {
	return verificationPrefix + token
}

// VerificationTokens provides Redis-backed storage for short-lived verification tokens.
type VerificationTokens struct {
	*sum.Store[models.VerificationToken]
}

// NewVerificationTokens creates a new verification tokens store backed by a Redis key-value provider.
func NewVerificationTokens(provider grub.StoreProvider) (*VerificationTokens, error) {
	store, err := sum.NewStore[models.VerificationToken](provider, "verification_tokens")
	if err != nil {
		return nil, err
	}
	return &VerificationTokens{Store: store}, nil
}

// Get retrieves a verification token by its token string.
func (s *VerificationTokens) Get(ctx context.Context, token string) (*models.VerificationToken, error) {
	return s.Store.Get(ctx, verificationKey(token))
}

// Set stores a verification token with the given TTL.
// A TTL of 0 means no expiration.
func (s *VerificationTokens) Set(ctx context.Context, token *models.VerificationToken, ttl time.Duration) error {
	return s.Store.Set(ctx, verificationKey(token.Token), token, ttl)
}

// Delete removes a verification token by its token string.
func (s *VerificationTokens) Delete(ctx context.Context, token string) error {
	return s.Store.Delete(ctx, verificationKey(token))
}
