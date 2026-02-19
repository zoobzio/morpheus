package config

import "github.com/zoobzio/check"

// Google holds configuration for Google OAuth integration.
type Google struct {
	ClientID     string `env:"MORPHEUS_GOOGLE_CLIENT_ID"`
	ClientSecret string `env:"MORPHEUS_GOOGLE_CLIENT_SECRET"`
	CallbackURL  string `env:"MORPHEUS_GOOGLE_CALLBACK_URL"`
}

// Validate validates the Google configuration.
func (c Google) Validate() error {
	return check.All(
		check.Str(c.ClientID, "client_id").Required().V(),
		check.Str(c.ClientSecret, "client_secret").Required().V(),
		check.Str(c.CallbackURL, "callback_url").Required().V(),
	).Err()
}
