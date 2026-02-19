package models

import "github.com/zoobzio/sum"

// RegisterBoundaries registers all model boundaries with the service registry.
// Must be called before sum.Freeze(k) in main.go.
func RegisterBoundaries(k sum.Key) error {
	if _, err := sum.NewBoundary[Provider](k); err != nil {
		return err
	}
	return nil
}
