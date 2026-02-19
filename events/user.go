package events

import (
	"github.com/zoobzio/capitan"
	"github.com/zoobzio/sum"
)

// UserEvent carries user lifecycle data.
type UserEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
}

// User signals.
var (
	UserCreatedSignal = capitan.NewSignal("morpheus.user.created", "User created")
	UserDeletedSignal = capitan.NewSignal("morpheus.user.deleted", "User deleted")
)

// User provides access to user lifecycle events.
var User = struct {
	Created sum.Event[UserEvent]
	Deleted sum.Event[UserEvent]
}{
	Created: sum.NewInfoEvent[UserEvent](UserCreatedSignal),
	Deleted: sum.NewInfoEvent[UserEvent](UserDeletedSignal),
}
