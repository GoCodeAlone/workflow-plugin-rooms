// Package rooms provides the workflow-plugin-rooms SDK plugin.
// It manages connection rooms: join, leave, broadcast, and member queries.
package rooms

import (
	"github.com/GoCodeAlone/workflow-plugin-rooms/internal"
	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// RoomHub is the public interface for room management.
type RoomHub = internal.RoomHub

// NewRoomsPlugin returns the rooms SDK plugin provider.
func NewRoomsPlugin() sdk.PluginProvider {
	return internal.NewRoomsPlugin()
}

// GetHub returns the global room hub once the rooms.manager module has initialized.
// Returns nil if the module has not started yet.
func GetHub() RoomHub {
	return internal.GetGlobalHub()
}

// NewHub creates a standalone room hub (useful for testing or custom wiring).
func NewHub() RoomHub {
	return internal.NewHub()
}
