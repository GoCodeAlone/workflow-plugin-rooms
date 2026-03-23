package rooms_test

import (
	"testing"

	"github.com/GoCodeAlone/workflow/wftest"
)

func TestRooms_JoinAndBroadcastPipeline(t *testing.T) {
	joinRec := wftest.RecordStep("step.room_join")
	joinRec.WithOutput(map[string]any{"joined": true, "room": "lobby"})

	broadcastRec := wftest.RecordStep("step.room_broadcast")
	broadcastRec.WithOutput(map[string]any{"recipients": 1})

	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  player-join:
    trigger:
      type: manual
    steps:
      - name: join
        type: step.room_join
        config:
          room: "lobby"
      - name: announce
        type: step.room_broadcast
        config:
          room: "lobby"
          message: "Player joined!"
`), joinRec, broadcastRec)

	result := h.ExecutePipeline("player-join", map[string]any{
		"connectionId": "conn-123",
		"room":         "lobby",
	})
	if result.Error != nil {
		t.Fatalf("pipeline failed: %v", result.Error)
	}
	if joinRec.CallCount() != 1 {
		t.Errorf("join step called %d times, want 1", joinRec.CallCount())
	}
	if broadcastRec.CallCount() != 1 {
		t.Errorf("broadcast step called %d times, want 1", broadcastRec.CallCount())
	}
}

func TestRooms_LeaveAndMembersPipeline(t *testing.T) {
	leaveRec := wftest.RecordStep("step.room_leave")
	leaveRec.WithOutput(map[string]any{"left": true, "room": "game-1"})

	membersRec := wftest.RecordStep("step.room_members")
	membersRec.WithOutput(map[string]any{"connections": []string{"conn-2", "conn-3"}, "count": 2})

	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  player-leave:
    trigger:
      type: manual
    steps:
      - name: leave
        type: step.room_leave
        config:
          room: "game-1"
      - name: list-members
        type: step.room_members
        config:
          room: "game-1"
`), leaveRec, membersRec)

	result := h.ExecutePipeline("player-leave", map[string]any{
		"connectionId": "conn-1",
		"room":         "game-1",
	})
	if result.Error != nil {
		t.Fatalf("pipeline failed: %v", result.Error)
	}
	if leaveRec.CallCount() != 1 {
		t.Errorf("leave step called %d times, want 1", leaveRec.CallCount())
	}
	if membersRec.CallCount() != 1 {
		t.Errorf("members step called %d times, want 1", membersRec.CallCount())
	}
}

func TestRooms_AllStepsPipeline(t *testing.T) {
	joinRec := wftest.RecordStep("step.room_join")
	joinRec.WithOutput(map[string]any{"joined": true})

	broadcastRec := wftest.RecordStep("step.room_broadcast")
	broadcastRec.WithOutput(map[string]any{"recipients": 3})

	membersRec := wftest.RecordStep("step.room_members")
	membersRec.WithOutput(map[string]any{"connections": []string{"conn-1", "conn-2", "conn-3"}, "count": 3})

	leaveRec := wftest.RecordStep("step.room_leave")
	leaveRec.WithOutput(map[string]any{"left": true})

	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  room-lifecycle:
    trigger:
      type: manual
    steps:
      - name: join
        type: step.room_join
        config:
          room: "arena"
      - name: broadcast-welcome
        type: step.room_broadcast
        config:
          room: "arena"
          message: "Welcome!"
      - name: list-members
        type: step.room_members
        config:
          room: "arena"
      - name: leave
        type: step.room_leave
        config:
          room: "arena"
`), joinRec, broadcastRec, membersRec, leaveRec)

	result := h.ExecutePipeline("room-lifecycle", map[string]any{
		"connectionId": "conn-42",
		"room":         "arena",
	})
	if result.Error != nil {
		t.Fatalf("pipeline failed: %v", result.Error)
	}

	for name, rec := range map[string]*wftest.Recorder{
		"join":      joinRec,
		"broadcast": broadcastRec,
		"members":   membersRec,
		"leave":     leaveRec,
	} {
		if rec.CallCount() != 1 {
			t.Errorf("%s step called %d times, want 1", name, rec.CallCount())
		}
	}
}
