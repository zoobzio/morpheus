package handlers

import "github.com/zoobzio/rocco"

// All returns all admin API endpoints for registration with the router.
func All() []rocco.Endpoint {
	return []rocco.Endpoint{
		// Users
		ListUsers,
		GetUser,
		DeleteUser,

		// Sessions
		ListSessions,
		RevokeSession,
	}
}
