package internal

import (
	"context"

	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

type roomLeaveStep struct{ name string }

func newRoomLeaveStep(name string, _ map[string]any) (sdk.StepInstance, error) {
	return &roomLeaveStep{name: name}, nil
}

func (s *roomLeaveStep) Execute(_ context.Context, _ map[string]any,
	_ map[string]map[string]any, current map[string]any,
	_ map[string]any, config map[string]any) (*sdk.StepResult, error) {

	h := GetGlobalHub()
	if h == nil {
		return &sdk.StepResult{Output: map[string]any{"error": "rooms.manager not initialized", "left": false}}, nil
	}

	connID, _ := config["connectionId"].(string)
	if connID == "" {
		connID, _ = current["connectionId"].(string)
	}
	if connID == "" {
		connID, _ = current["connID"].(string)
	}

	room, _ := config["room"].(string)

	if connID == "" || room == "" {
		return &sdk.StepResult{Output: map[string]any{"error": "connectionId and room are required", "left": false}}, nil
	}

	h.LeaveRoom(room, connID)
	return &sdk.StepResult{Output: map[string]any{"left": true, "room": room}}, nil
}
