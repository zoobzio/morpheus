package config

import (
	"fmt"

	"github.com/zoobzio/check"
)

// Database holds configuration for the PostgreSQL database connection.
type Database struct {
	Host     string `env:"MORPHEUS_DB_HOST" default:"localhost"`
	Port     int    `env:"MORPHEUS_DB_PORT" default:"5432"`
	User     string `env:"MORPHEUS_DB_USER" default:"morpheus"`
	Password string `env:"MORPHEUS_DB_PASSWORD"`
	Name     string `env:"MORPHEUS_DB_NAME" default:"morpheus"`
	SSLMode  string `env:"MORPHEUS_DB_SSLMODE" default:"disable"`
}

// Validate validates the Database configuration.
func (c Database) Validate() error {
	return check.All(
		check.Str(c.Host, "host").Required().MaxLen(255).V(),
		check.Int(c.Port, "port").Positive().Max(65535).V(),
		check.Str(c.User, "user").Required().MaxLen(255).V(),
		check.Str(c.Name, "name").Required().MaxLen(255).V(),
		check.Str(c.SSLMode, "ssl_mode").Required().OneOf([]string{"disable", "require", "verify-ca", "verify-full"}).V(),
	).Err()
}

// DSN returns the PostgreSQL connection string.
func (c Database) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}
