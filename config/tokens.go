package config

import "time"

// Tokens holds TTL configuration for verification token flows.
type Tokens struct {
	EmailVerifyTTL  time.Duration `env:"MORPHEUS_TOKEN_EMAIL_VERIFY_TTL" default:"24h"`
	MagicLinkTTL    time.Duration `env:"MORPHEUS_TOKEN_MAGIC_LINK_TTL" default:"15m"`
	PasswordResetTTL time.Duration `env:"MORPHEUS_TOKEN_PASSWORD_RESET_TTL" default:"1h"`
}
