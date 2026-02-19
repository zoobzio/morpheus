package wire

import "github.com/zoobzio/sum"

// RegisterBoundaries creates and registers all public API wire boundaries.
// Must be called before sum.Freeze(k) in cmd/app/main.go.
func RegisterBoundaries(k sum.Key) error {
	if _, err := sum.NewBoundary[UserResponse](k); err != nil {
		return err
	}
	return nil
}
