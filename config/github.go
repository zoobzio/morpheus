package config

import "github.com/zoobzio/check"

// GitHub holds configuration for GitHub OAuth integration.
type GitHub struct {
	ClientID     string `env:"MORPHEUS_GITHUB_CLIENT_ID"`
	ClientSecret string `env:"MORPHEUS_GITHUB_CLIENT_SECRET"`
	CallbackURL  string `env:"MORPHEUS_GITHUB_CALLBACK_URL"`
}

// Validate validates the GitHub configuration.
func (c GitHub) Validate() error {
	return check.All(
		check.Str(c.ClientID, "client_id").Required().V(),
		check.Str(c.ClientSecret, "client_secret").Required().V(),
		check.Str(c.CallbackURL, "callback_url").Required().V(),
	).Err()
}
