package internal

import (
	"context"

	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

type roomMembersStep struct{ name string }

func newRoomMembersStep(name string, _ map[string]any) (sdk.StepInstance, error) {
	return &roomMembersStep{name: name}, nil
}

func (s *roomMembersStep) Execute(_ context.Context, _ map[string]any,
	_ map[string]map[string]any, _ map[string]any,
	_ map[string]any, config map[string]any) (*sdk.StepResult, error) {

	h := GetGlobalHub()
	if h == nil {
		return &sdk.StepResult{Output: map[string]any{"error": "rooms.manager not initialized", "connections": []string{}}}, nil
	}

	room, _ := config["room"].(string)
	members := h.GetMembers(room)
	if members == nil {
		members = []string{}
	}

	return &sdk.StepResult{Output: map[string]any{"connections": members, "count": len(members)}}, nil
}
