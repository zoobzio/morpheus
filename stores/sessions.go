package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/zoobzio/grub"
	"github.com/zoobzio/sum"
	"github.com/zoobzio/sumatra/models"
)

const (
	sessionPrefix   = "session:"
	userIndexPrefix = "user_sessions:"
)

// sessionKey returns the primary key for a session token.
func sessionKey(token string) string {
	return sessionPrefix + token
}

// userIndexKey returns the index key for a user's session entry.
func userIndexKey(userID, token string) string {
	return fmt.Sprintf("%s%s:%s", userIndexPrefix, userID, token)
}

// userIndexScanPrefix returns the prefix used to list all sessions for a user.
func userIndexScanPrefix(userID string) string {
	return userIndexPrefix + userID + ":"
}

// Sessions provides Redis-backed session storage with TTL and user index support.
type Sessions struct {
	*sum.Store[models.Session]
}

// NewSessions creates a new sessions store backed by a Redis key-value provider.
func NewSessions(provider grub.StoreProvider) (*Sessions, error) {
	store, err := sum.NewStore[models.Session](provider, "sessions")
	if err != nil {
		return nil, err
	}
	return &Sessions{Store: store}, nil
}

// Get retrieves a session by its token.
func (s *Sessions) Get(ctx context.Context, token string) (*models.Session, error) {
	return s.Store.Get(ctx, sessionKey(token))
}

// Set stores a session with the given TTL.
// A TTL of 0 means no expiration.
func (s *Sessions) Set(ctx context.Context, session *models.Session, ttl time.Duration) error {
	return s.Store.Set(ctx, sessionKey(session.Token), session, ttl)
}

// Delete removes a session by its token.
func (s *Sessions) Delete(ctx context.Context, token string) error {
	return s.Store.Delete(ctx, sessionKey(token))
}

// Exists reports whether a session exists for the given token.
func (s *Sessions) Exists(ctx context.Context, token string) (bool, error) {
	return s.Store.Exists(ctx, sessionKey(token))
}

// SetWithUserIndex stores a session and writes a corresponding user index entry.
// The user index entry enables ListByUser and DeleteByUser operations.
// A TTL of 0 means no expiration.
func (s *Sessions) SetWithUserIndex(ctx context.Context, session *models.Session, ttl time.Duration) error {
	if err := s.Store.Set(ctx, sessionKey(session.Token), session, ttl); err != nil {
		return err
	}
	// Store an index entry: user_sessions:{userID}:{token} → token
	// The value is a minimal session stub used only as a marker; the real session
	// is looked up by token. We reuse models.Session with only the token set.
	marker := &models.Session{
		Token:  session.Token,
		UserID: session.UserID,
	}
	return s.Store.Set(ctx, userIndexKey(session.UserID, session.Token), marker, ttl)
}

// ListByUser returns up to limit session tokens belonging to userID.
// It scans the user index and returns the token values.
func (s *Sessions) ListByUser(ctx context.Context, userID string, limit int) ([]string, error) {
	keys, err := s.Store.List(ctx, userIndexScanPrefix(userID), limit)
	if err != nil {
		return nil, err
	}
	tokens := make([]string, 0, len(keys))
	for _, key := range keys {
		entry, err := s.Store.Get(ctx, key)
		if err != nil || entry == nil {
			continue
		}
		tokens = append(tokens, entry.Token)
	}
	return tokens, nil
}

// DeleteByUser revokes all sessions for userID by scanning the user index
// and deleting each session and its index entry.
func (s *Sessions) DeleteByUser(ctx context.Context, userID string) error {
	keys, err := s.Store.List(ctx, userIndexScanPrefix(userID), 0)
	if err != nil {
		return err
	}
	for _, key := range keys {
		entry, err := s.Store.Get(ctx, key)
		if err != nil || entry == nil {
			// Best-effort: remove the index entry even if session is missing.
			_ = s.Store.Delete(ctx, key)
			continue
		}
		_ = s.Store.Delete(ctx, sessionKey(entry.Token))
		_ = s.Store.Delete(ctx, key)
	}
	return nil
}
