//go:build testing

package testing

import (
	"context"
	"time"

	apicontracts "github.com/zoobzio/sumatra/api/contracts"
	admincontracts "github.com/zoobzio/sumatra/admin/contracts"
	"github.com/zoobzio/sumatra/models"
)

// Compile-time interface checks.
var (
	_ apicontracts.Users     = (*MockAPIUsers)(nil)
	_ apicontracts.Providers = (*MockAPIProviders)(nil)
	_ apicontracts.Sessions  = (*MockAPISessions)(nil)

	_ admincontracts.Users     = (*MockAdminUsers)(nil)
	_ admincontracts.Sessions  = (*MockAdminSessions)(nil)
	_ admincontracts.Providers = (*MockAdminProviders)(nil)
)

// MockAPIUsers is a mock implementation of api/contracts.Users.
type MockAPIUsers struct {
	OnGet          func(ctx context.Context, key string) (*models.User, error)
	OnGetByEmail   func(ctx context.Context, email string) (*models.User, error)
	OnSet          func(ctx context.Context, key string, user *models.User) error
}

func (m *MockAPIUsers) Get(ctx context.Context, key string) (*models.User, error) {
	if m.OnGet != nil {
		return m.OnGet(ctx, key)
	}
	return &models.User{}, nil
}

func (m *MockAPIUsers) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.OnGetByEmail != nil {
		return m.OnGetByEmail(ctx, email)
	}
	return &models.User{}, nil
}

func (m *MockAPIUsers) Set(ctx context.Context, key string, user *models.User) error {
	if m.OnSet != nil {
		return m.OnSet(ctx, key, user)
	}
	return nil
}

// MockAPIProviders is a mock implementation of api/contracts.Providers.
type MockAPIProviders struct {
	OnGetByProviderUser func(ctx context.Context, providerType models.ProviderType, providerUserID string) (*models.Provider, error)
	OnSet               func(ctx context.Context, key string, provider *models.Provider) error
}

func (m *MockAPIProviders) GetByProviderUser(ctx context.Context, providerType models.ProviderType, providerUserID string) (*models.Provider, error) {
	if m.OnGetByProviderUser != nil {
		return m.OnGetByProviderUser(ctx, providerType, providerUserID)
	}
	return &models.Provider{}, nil
}

func (m *MockAPIProviders) Set(ctx context.Context, key string, provider *models.Provider) error {
	if m.OnSet != nil {
		return m.OnSet(ctx, key, provider)
	}
	return nil
}

// MockAPISessions is a mock implementation of api/contracts.Sessions.
type MockAPISessions struct {
	OnGet              func(ctx context.Context, token string) (*models.Session, error)
	OnSetWithUserIndex func(ctx context.Context, session *models.Session, ttl time.Duration) error
	OnDelete           func(ctx context.Context, token string) error
}

func (m *MockAPISessions) Get(ctx context.Context, token string) (*models.Session, error) {
	if m.OnGet != nil {
		return m.OnGet(ctx, token)
	}
	return &models.Session{}, nil
}

func (m *MockAPISessions) SetWithUserIndex(ctx context.Context, session *models.Session, ttl time.Duration) error {
	if m.OnSetWithUserIndex != nil {
		return m.OnSetWithUserIndex(ctx, session, ttl)
	}
	return nil
}

func (m *MockAPISessions) Delete(ctx context.Context, token string) error {
	if m.OnDelete != nil {
		return m.OnDelete(ctx, token)
	}
	return nil
}

// MockAdminUsers is a mock implementation of admin/contracts.Users.
type MockAdminUsers struct {
	OnGet    func(ctx context.Context, key string) (*models.User, error)
	OnList   func(ctx context.Context, limit, offset int) ([]*models.User, error)
	OnCount  func(ctx context.Context) (float64, error)
	OnDelete func(ctx context.Context, key string) error
}

func (m *MockAdminUsers) Get(ctx context.Context, key string) (*models.User, error) {
	if m.OnGet != nil {
		return m.OnGet(ctx, key)
	}
	return &models.User{}, nil
}

func (m *MockAdminUsers) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	if m.OnList != nil {
		return m.OnList(ctx, limit, offset)
	}
	return nil, nil
}

func (m *MockAdminUsers) Count(ctx context.Context) (float64, error) {
	if m.OnCount != nil {
		return m.OnCount(ctx)
	}
	return 0, nil
}

func (m *MockAdminUsers) Delete(ctx context.Context, key string) error {
	if m.OnDelete != nil {
		return m.OnDelete(ctx, key)
	}
	return nil
}

// MockAdminSessions is a mock implementation of admin/contracts.Sessions.
type MockAdminSessions struct {
	OnGet          func(ctx context.Context, token string) (*models.Session, error)
	OnDelete       func(ctx context.Context, token string) error
	OnListByUser   func(ctx context.Context, userID string, limit int) ([]string, error)
	OnDeleteByUser func(ctx context.Context, userID string) error
}

func (m *MockAdminSessions) Get(ctx context.Context, token string) (*models.Session, error) {
	if m.OnGet != nil {
		return m.OnGet(ctx, token)
	}
	return &models.Session{}, nil
}

func (m *MockAdminSessions) Delete(ctx context.Context, token string) error {
	if m.OnDelete != nil {
		return m.OnDelete(ctx, token)
	}
	return nil
}

func (m *MockAdminSessions) ListByUser(ctx context.Context, userID string, limit int) ([]string, error) {
	if m.OnListByUser != nil {
		return m.OnListByUser(ctx, userID, limit)
	}
	return nil, nil
}

func (m *MockAdminSessions) DeleteByUser(ctx context.Context, userID string) error {
	if m.OnDeleteByUser != nil {
		return m.OnDeleteByUser(ctx, userID)
	}
	return nil
}

// MockAdminProviders is a mock implementation of admin/contracts.Providers.
type MockAdminProviders struct {
	OnDeleteByUser func(ctx context.Context, userID string) error
}

func (m *MockAdminProviders) DeleteByUser(ctx context.Context, userID string) error {
	if m.OnDeleteByUser != nil {
		return m.OnDeleteByUser(ctx, userID)
	}
	return nil
}
