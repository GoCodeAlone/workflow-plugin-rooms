package internal

import (
	"context"

	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

type roomJoinStep struct{ name string }

func newRoomJoinStep(name string, _ map[string]any) (sdk.StepInstance, error) {
	return &roomJoinStep{name: name}, nil
}

func (s *roomJoinStep) Execute(_ context.Context, _ map[string]any,
	_ map[string]map[string]any, current map[string]any,
	_ map[string]any, config map[string]any) (*sdk.StepResult, error) {

	h := GetGlobalHub()
	if h == nil {
		return &sdk.StepResult{Output: map[string]any{"error": "rooms.manager not initialized", "joined": false}}, nil
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
		return &sdk.StepResult{Output: map[string]any{"error": "connectionId and room are required", "joined": false}}, nil
	}

	joined := h.JoinRoom(room, connID)
	if !joined {
		return &sdk.StepResult{Output: map[string]any{"error": "connection not found", "joined": false}}, nil
	}
	return &sdk.StepResult{Output: map[string]any{"joined": true, "room": room}}, nil
}
