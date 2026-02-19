// Package identity implements the aegis IdentityService for mesh consumers.
package identity

import (
	"context"

	"github.com/zoobzio/aegis/proto/identity"
	"github.com/zoobzio/sumatra/api/contracts"
)

// Server implements identity.IdentityServiceServer.
type Server struct {
	identity.UnimplementedIdentityServiceServer
	users     contracts.Users
	sessions  contracts.Sessions
	providers contracts.Providers
}

// New creates a new identity server.
func New(users contracts.Users, sessions contracts.Sessions, providers contracts.Providers) *Server {
	return &Server{
		users:     users,
		sessions:  sessions,
		providers: providers,
	}
}

// ValidateSession checks if a session token is valid.
func (s *Server) ValidateSession(ctx context.Context, req *identity.ValidateSessionRequest) (*identity.ValidateSessionResponse, error) {
	session, err := s.sessions.Get(ctx, req.Token)
	if err != nil || session == nil {
		return &identity.ValidateSessionResponse{Valid: false}, nil
	}
	return &identity.ValidateSessionResponse{
		Valid:     !session.IsExpired(),
		UserId:    session.UserID,
		ExpiresAt: session.ExpiresAt.Unix(),
	}, nil
}

// GetUser retrieves a user by ID.
func (s *Server) GetUser(ctx context.Context, req *identity.GetUserRequest) (*identity.GetUserResponse, error) {
	user, err := s.users.Get(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	return userToProto(user), nil
}

// GetUserByEmail retrieves a user by email address.
func (s *Server) GetUserByEmail(ctx context.Context, req *identity.GetUserByEmailRequest) (*identity.GetUserResponse, error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	return userToProto(user), nil
}

// ListProviders returns the OAuth providers linked to a user.
func (s *Server) ListProviders(ctx context.Context, req *identity.ListProvidersRequest) (*identity.ListProvidersResponse, error) {
	providers, err := s.providers.ListByUser(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	resp := &identity.ListProvidersResponse{
		Providers: make([]*identity.Provider, len(providers)),
	}
	for i, p := range providers {
		resp.Providers[i] = &identity.Provider{
			Type:           string(p.Type),
			ProviderUserId: p.ProviderUserID,
			LinkedAt:       p.CreatedAt.Unix(),
		}
	}
	return resp, nil
}
