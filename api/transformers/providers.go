package transformers

import (
	"github.com/zoobzio/sumatra/api/wire"
	"github.com/zoobzio/sumatra/models"
)

// ProviderToResponse transforms a Provider model to a public API ProviderResponse.
func ProviderToResponse(p *models.Provider) wire.ProviderResponse {
	return wire.ProviderResponse{
		Type:           string(p.Type),
		ProviderUserID: p.ProviderUserID,
		LinkedAt:       p.CreatedAt,
	}
}

// ProvidersToList transforms a slice of Provider models to a public API ProviderListResponse.
func ProvidersToList(providers []*models.Provider) wire.ProviderListResponse {
	resp := wire.ProviderListResponse{
		Providers: make([]wire.ProviderResponse, len(providers)),
	}
	for i, p := range providers {
		resp.Providers[i] = ProviderToResponse(p)
	}
	return resp
}
