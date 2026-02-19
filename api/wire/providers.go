package wire

import "time"

// ProviderResponse is the public API response for a single linked OAuth provider.
type ProviderResponse struct {
	Type           string    `json:"type" description:"OAuth provider type" example:"github"`
	ProviderUserID string    `json:"provider_user_id" description:"Provider-side user ID" example:"12345678"`
	LinkedAt       time.Time `json:"linked_at" description:"Time the provider was linked"`
}

// Clone returns a deep copy of ProviderResponse.
func (p ProviderResponse) Clone() ProviderResponse {
	return p
}

// ProviderListResponse is the public API response for listing linked OAuth providers.
type ProviderListResponse struct {
	Providers []ProviderResponse `json:"providers" description:"Linked OAuth providers"`
}

// Clone returns a deep copy of ProviderListResponse.
func (p ProviderListResponse) Clone() ProviderListResponse {
	c := p
	if p.Providers != nil {
		c.Providers = make([]ProviderResponse, len(p.Providers))
		copy(c.Providers, p.Providers)
	}
	return c
}
