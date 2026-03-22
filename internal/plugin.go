package internal

import (
	"fmt"
	"sync"

	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

var (
	globalHub   *Hub
	globalHubMu sync.RWMutex
)

// GetGlobalHub returns the global room hub once the rooms.manager module has initialized.
func GetGlobalHub() *Hub {
	globalHubMu.RLock()
	defer globalHubMu.RUnlock()
	return globalHub
}

// SetGlobalHub sets the global hub (called by the module on Init).
func SetGlobalHub(h *Hub) {
	globalHubMu.Lock()
	globalHub = h
	globalHubMu.Unlock()
}

type roomsPlugin struct{}

// NewRoomsPlugin returns the rooms SDK plugin provider.
func NewRoomsPlugin() sdk.PluginProvider {
	return &roomsPlugin{}
}

func (p *roomsPlugin) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "workflow-plugin-rooms",
		Version:     "0.1.0",
		Author:      "GoCodeAlone",
		Description: "Room management for workflow applications — join, leave, broadcast, members",
	}
}

func (p *roomsPlugin) ModuleTypes() []string {
	return []string{"rooms.manager"}
}

func (p *roomsPlugin) StepTypes() []string {
	return []string{
		"step.room_join",
		"step.room_leave",
		"step.room_broadcast",
		"step.room_members",
	}
}

func (p *roomsPlugin) CreateModule(typeName, name string, config map[string]any) (sdk.ModuleInstance, error) {
	switch typeName {
	case "rooms.manager":
		return newRoomsManagerModule(name, config)
	default:
		return nil, fmt.Errorf("unknown module type %q", typeName)
	}
}

func (p *roomsPlugin) CreateStep(typeName, name string, config map[string]any) (sdk.StepInstance, error) {
	switch typeName {
	case "step.room_join":
		return newRoomJoinStep(name, config)
	case "step.room_leave":
		return newRoomLeaveStep(name, config)
	case "step.room_broadcast":
		return newRoomBroadcastStep(name, config)
	case "step.room_members":
		return newRoomMembersStep(name, config)
	default:
		return nil, fmt.Errorf("unknown step type %q", typeName)
	}
}
