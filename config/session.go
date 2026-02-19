package config

import (
	"time"

	"github.com/zoobzio/check"
)

// Session holds configuration for session management and cookie settings.
type Session struct {
	TTL          time.Duration `env:"MORPHEUS_SESSION_TTL" default:"168h"`
	CookieName   string        `env:"MORPHEUS_SESSION_COOKIE_NAME" default:"session"`
	CookieDomain string        `env:"MORPHEUS_SESSION_COOKIE_DOMAIN"`
	CookieSecure bool          `env:"MORPHEUS_SESSION_COOKIE_SECURE"`
	CookiePath   string        `env:"MORPHEUS_SESSION_COOKIE_PATH" default:"/"`
	StateSecret  string        `env:"MORPHEUS_SESSION_STATE_SECRET"`
}

// Validate validates the Session configuration.
func (c Session) Validate() error {
	return check.All(
		check.Str(c.CookieName, "cookie_name").Required().V(),
		check.Str(c.CookiePath, "cookie_path").Required().V(),
		check.Str(c.StateSecret, "state_secret").Required().MinLen(32).V(),
	).Err()
}
