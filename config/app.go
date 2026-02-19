package config

import (
	"fmt"

	"github.com/zoobzio/check"
)

// App holds configuration for the application server.
type App struct {
	Port        int    `env:"MORPHEUS_PORT" default:"8080"`
	Environment string `env:"MORPHEUS_ENV" default:"development"`
}

// Validate validates the App configuration.
func (c App) Validate() error {
	return check.All(
		check.Int(c.Port, "port").Positive().Max(65535).V(),
		check.Str(c.Environment, "environment").Required().V(),
	).Err()
}

// Addr returns the listen address for the HTTP server.
func (c App) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

// IsDevelopment reports whether the application is running in development mode.
func (c App) IsDevelopment() bool {
	return c.Environment == "development"
}
