package models

import (
	"testing"
)

func TestProvider_Validate_Success(t *testing.T) {
	p := Provider{
		UserID:         "01942d3a-1234-7abc-8def-0123456789ab",
		Type:           ProviderTypeGitHub,
		ProviderUserID: "583231",
		AccessToken:    "gho_test_access_token",
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestProvider_Validate_MissingUserID(t *testing.T) {
	p := Provider{
		Type:           ProviderTypeGitHub,
		ProviderUserID: "583231",
		AccessToken:    "gho_test_access_token",
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing UserID, got nil")
	}
}

func TestProvider_Validate_MissingType(t *testing.T) {
	p := Provider{
		UserID:         "01942d3a-1234-7abc-8def-0123456789ab",
		ProviderUserID: "583231",
		AccessToken:    "gho_test_access_token",
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing Type, got nil")
	}
}

func TestProvider_Validate_InvalidType(t *testing.T) {
	p := Provider{
		UserID:         "01942d3a-1234-7abc-8def-0123456789ab",
		Type:           ProviderType("gitlab"),
		ProviderUserID: "583231",
		AccessToken:    "gho_test_access_token",
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for unsupported provider type, got nil")
	}
}

func TestProvider_Validate_MissingProviderUserID(t *testing.T) {
	p := Provider{
		UserID:      "01942d3a-1234-7abc-8def-0123456789ab",
		Type:        ProviderTypeGitHub,
		AccessToken: "gho_test_access_token",
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing ProviderUserID, got nil")
	}
}

func TestProvider_Validate_MissingAccessToken(t *testing.T) {
	p := Provider{
		UserID:         "01942d3a-1234-7abc-8def-0123456789ab",
		Type:           ProviderTypeGitHub,
		ProviderUserID: "583231",
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing AccessToken, got nil")
	}
}

func TestProvider_Clone(t *testing.T) {
	p := Provider{
		ID:             42,
		UserID:         "01942d3a-1234-7abc-8def-0123456789ab",
		Type:           ProviderTypeGitHub,
		ProviderUserID: "583231",
		AccessToken:    "gho_test_access_token",
	}
	c := p.Clone()

	if c.ID != p.ID {
		t.Errorf("ID mismatch: got %d want %d", c.ID, p.ID)
	}
	if c.UserID != p.UserID {
		t.Errorf("UserID mismatch: got %q want %q", c.UserID, p.UserID)
	}
	if c.Type != p.Type {
		t.Errorf("Type mismatch: got %q want %q", c.Type, p.Type)
	}
	if c.ProviderUserID != p.ProviderUserID {
		t.Errorf("ProviderUserID mismatch: got %q want %q", c.ProviderUserID, p.ProviderUserID)
	}
	if c.AccessToken != p.AccessToken {
		t.Errorf("AccessToken mismatch: got %q want %q", c.AccessToken, p.AccessToken)
	}
}

func TestProvider_Clone_Independence(t *testing.T) {
	p := Provider{
		UserID:         "01942d3a-1234-7abc-8def-0123456789ab",
		Type:           ProviderTypeGitHub,
		ProviderUserID: "583231",
		AccessToken:    "gho_test_access_token",
	}
	c := p.Clone()

	// Mutate the clone; original must be unaffected.
	c.AccessToken = "mutated"

	if p.AccessToken != "gho_test_access_token" {
		t.Errorf("original AccessToken was mutated: got %q", p.AccessToken)
	}
}
