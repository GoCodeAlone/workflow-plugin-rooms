package internal

import (
	"context"

	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

type roomBroadcastStep struct{ name string }

func newRoomBroadcastStep(name string, _ map[string]any) (sdk.StepInstance, error) {
	return &roomBroadcastStep{name: name}, nil
}

func (s *roomBroadcastStep) Execute(_ context.Context, _ map[string]any,
	_ map[string]map[string]any, _ map[string]any,
	_ map[string]any, config map[string]any) (*sdk.StepResult, error) {

	h := GetGlobalHub()
	if h == nil {
		return &sdk.StepResult{Output: map[string]any{"error": "rooms.manager not initialized", "recipients": 0}}, nil
	}

	room, _ := config["room"].(string)
	message, _ := config["message"].(string)
	exclude, _ := config["exclude"].(string)

	count := h.BroadcastToRoom(room, []byte(message), exclude)
	return &sdk.StepResult{Output: map[string]any{"recipients": count}}, nil
}
