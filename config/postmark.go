package config

import "github.com/zoobzio/check"

// Postmark holds configuration for the Postmark transactional email API.
type Postmark struct {
	ServerToken string `env:"MORPHEUS_POSTMARK_SERVER_TOKEN"`
	DefaultFrom string `env:"MORPHEUS_POSTMARK_DEFAULT_FROM"`
}

// Validate validates the Postmark configuration.
func (c Postmark) Validate() error {
	return check.All(
		check.Str(c.ServerToken, "server_token").Required().V(),
		check.Str(c.DefaultFrom, "default_from").Required().V(),
	).Err()
}
