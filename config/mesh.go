package config

import (
	"fmt"

	"github.com/zoobzio/check"
)

// Mesh holds configuration for the aegis mesh node.
type Mesh struct {
	ID      string `env:"MORPHEUS_MESH_ID" default:"morpheus-1"`
	Name    string `env:"MORPHEUS_MESH_NAME" default:"Morpheus"`
	Host    string `env:"MORPHEUS_MESH_HOST" default:"localhost"`
	Port    int    `env:"MORPHEUS_MESH_PORT" default:"8443"`
	CertDir string `env:"MORPHEUS_MESH_CERT_DIR" default:"./certs"`
}

// Validate validates the Mesh configuration.
func (c Mesh) Validate() error {
	return check.All(
		check.Str(c.ID, "id").Required().V(),
		check.Str(c.Name, "name").Required().V(),
		check.Int(c.Port, "port").Positive().Max(65535).V(),
		check.Str(c.CertDir, "cert_dir").Required().V(),
	).Err()
}

// Addr returns the gRPC listen address.
func (c Mesh) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
