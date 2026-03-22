package internal

import (
	"context"

	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

type roomsManagerModule struct {
	name string
	hub  *Hub
}

func newRoomsManagerModule(name string, _ map[string]any) (sdk.ModuleInstance, error) {
	return &roomsManagerModule{name: name}, nil
}

func (m *roomsManagerModule) Init() error {
	m.hub = NewHub()
	SetGlobalHub(m.hub)
	return nil
}

func (m *roomsManagerModule) Start(_ context.Context) error {
	return nil
}

func (m *roomsManagerModule) Stop(_ context.Context) error {
	globalHubMu.Lock()
	if globalHub == m.hub {
		globalHub = nil
	}
	globalHubMu.Unlock()
	return nil
}
