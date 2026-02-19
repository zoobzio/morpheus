package events

import (
	"github.com/zoobzio/capitan"
	"github.com/zoobzio/sum"
)

// SessionEvent carries session lifecycle data.
type SessionEvent struct {
	UserID int64 `json:"user_id"`
}

// Session signals.
var (
	SessionCreatedSignal = capitan.NewSignal("morpheus.session.created", "Session created")
	SessionRevokedSignal = capitan.NewSignal("morpheus.session.revoked", "Session revoked")
)

// Session provides access to session lifecycle events.
var Session = struct {
	Created sum.Event[SessionEvent]
	Revoked sum.Event[SessionEvent]
}{
	Created: sum.NewInfoEvent[SessionEvent](SessionCreatedSignal),
	Revoked: sum.NewInfoEvent[SessionEvent](SessionRevokedSignal),
}
