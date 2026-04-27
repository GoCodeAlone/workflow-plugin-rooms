package internal

import (
	"context"
	"fmt"

	"github.com/GoCodeAlone/workflow-plugin-rooms/internal/contracts"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func typedRoomJoin(ctx context.Context, req sdk.TypedStepRequest[*contracts.RoomJoinConfig, *contracts.RoomJoinInput]) (*sdk.TypedStepResult[*contracts.RoomJoinOutput], error) {
	step, err := newRoomJoinStep("", nil)
	if err != nil {
		return nil, err
	}
	result, err := step.Execute(ctx, req.TriggerData, req.StepOutputs, req.Current, req.Metadata, map[string]any{
		"connectionId": firstNonEmpty(req.Input.GetConnectionId(), req.Config.GetConnectionId()),
		"room":         firstNonEmpty(req.Input.GetRoom(), req.Config.GetRoom()),
	})
	if err != nil {
		return nil, err
	}
	return &sdk.TypedStepResult[*contracts.RoomJoinOutput]{Output: roomJoinOutputFromMap(result.Output), StopPipeline: result.StopPipeline}, nil
}

func typedRoomLeave(ctx context.Context, req sdk.TypedStepRequest[*contracts.RoomLeaveConfig, *contracts.RoomLeaveInput]) (*sdk.TypedStepResult[*contracts.RoomLeaveOutput], error) {
	step, err := newRoomLeaveStep("", nil)
	if err != nil {
		return nil, err
	}
	result, err := step.Execute(ctx, req.TriggerData, req.StepOutputs, req.Current, req.Metadata, map[string]any{
		"connectionId": firstNonEmpty(req.Input.GetConnectionId(), req.Config.GetConnectionId()),
		"room":         firstNonEmpty(req.Input.GetRoom(), req.Config.GetRoom()),
	})
	if err != nil {
		return nil, err
	}
	return &sdk.TypedStepResult[*contracts.RoomLeaveOutput]{Output: roomLeaveOutputFromMap(result.Output), StopPipeline: result.StopPipeline}, nil
}

func typedRoomBroadcast(ctx context.Context, req sdk.TypedStepRequest[*contracts.RoomBroadcastConfig, *contracts.RoomBroadcastInput]) (*sdk.TypedStepResult[*contracts.RoomBroadcastOutput], error) {
	step, err := newRoomBroadcastStep("", nil)
	if err != nil {
		return nil, err
	}
	room := firstNonEmpty(req.Input.GetRoom(), req.Config.GetRoom())
	message := firstNonEmpty(req.Input.GetMessage(), req.Config.GetMessage())
	if err := requireNonEmpty("room", room); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("message", message); err != nil {
		return nil, err
	}
	result, err := step.Execute(ctx, req.TriggerData, req.StepOutputs, req.Current, req.Metadata, map[string]any{
		"room":    room,
		"message": message,
		"exclude": firstNonEmpty(req.Input.GetExclude(), req.Config.GetExclude()),
	})
	if err != nil {
		return nil, err
	}
	return &sdk.TypedStepResult[*contracts.RoomBroadcastOutput]{Output: roomBroadcastOutputFromMap(result.Output), StopPipeline: result.StopPipeline}, nil
}

func typedRoomMembers(ctx context.Context, req sdk.TypedStepRequest[*contracts.RoomMembersConfig, *contracts.RoomMembersInput]) (*sdk.TypedStepResult[*contracts.RoomMembersOutput], error) {
	step, err := newRoomMembersStep("", nil)
	if err != nil {
		return nil, err
	}
	room := firstNonEmpty(req.Input.GetRoom(), req.Config.GetRoom())
	if err := requireNonEmpty("room", room); err != nil {
		return nil, err
	}
	result, err := step.Execute(ctx, req.TriggerData, req.StepOutputs, req.Current, req.Metadata, map[string]any{
		"room": room,
	})
	if err != nil {
		return nil, err
	}
	return &sdk.TypedStepResult[*contracts.RoomMembersOutput]{Output: roomMembersOutputFromMap(result.Output), StopPipeline: result.StopPipeline}, nil
}

func requireNonEmpty(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func roomJoinOutputFromMap(values map[string]any) *contracts.RoomJoinOutput {
	return &contracts.RoomJoinOutput{
		Joined: boolValue(values["joined"]),
		Room:   stringValue(values["room"]),
		Error:  stringValue(values["error"]),
	}
}

func roomLeaveOutputFromMap(values map[string]any) *contracts.RoomLeaveOutput {
	return &contracts.RoomLeaveOutput{
		Left:  boolValue(values["left"]),
		Room:  stringValue(values["room"]),
		Error: stringValue(values["error"]),
	}
}

func roomBroadcastOutputFromMap(values map[string]any) *contracts.RoomBroadcastOutput {
	return &contracts.RoomBroadcastOutput{
		Recipients: int32Value(values["recipients"]),
		Error:      stringValue(values["error"]),
	}
}

func roomMembersOutputFromMap(values map[string]any) *contracts.RoomMembersOutput {
	connections, _ := values["connections"].([]string)
	return &contracts.RoomMembersOutput{
		Connections: append([]string(nil), connections...),
		Count:       int32Value(values["count"]),
		Error:       stringValue(values["error"]),
	}
}

func stringValue(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func boolValue(value any) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

func int32Value(value any) int32 {
	switch v := value.(type) {
	case int:
		return int32(v)
	case int32:
		return v
	case int64:
		return int32(v)
	case float64:
		return int32(v)
	default:
		return 0
	}
}
