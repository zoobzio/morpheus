package config

import (
	"fmt"

	"github.com/zoobzio/check"
)

// Redis holds configuration for the Redis connection.
type Redis struct {
	Host     string `env:"MORPHEUS_REDIS_HOST" default:"localhost"`
	Port     int    `env:"MORPHEUS_REDIS_PORT" default:"6379"`
	Password string `env:"MORPHEUS_REDIS_PASSWORD"`
	DB       int    `env:"MORPHEUS_REDIS_DB" default:"0"`
}

// Validate validates the Redis configuration.
func (c Redis) Validate() error {
	return check.All(
		check.Str(c.Host, "host").Required().MaxLen(255).V(),
		check.Int(c.Port, "port").Positive().Max(65535).V(),
		check.Int(c.DB, "db").NonNegative().V(),
	).Err()
}

// Addr returns the host:port address for the Redis connection.
func (c Redis) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
