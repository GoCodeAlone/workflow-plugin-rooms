package internal

import (
	"context"
	"testing"
)

func setupTestHub(t *testing.T) (*Hub, func()) {
	t.Helper()
	h := NewHub()

	globalHubMu.Lock()
	prev := globalHub
	globalHub = h
	globalHubMu.Unlock()

	return h, func() {
		globalHubMu.Lock()
		globalHub = prev
		globalHubMu.Unlock()
	}
}

func TestRoomJoinStep(t *testing.T) {
	h, cleanup := setupTestHub(t)
	defer cleanup()

	h.RegisterConnection("conn-1")

	step, err := newRoomJoinStep("join", nil)
	if err != nil {
		t.Fatal(err)
	}

	result, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"connectionId": "conn-1", "room": "lobby"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output["joined"] != true {
		t.Fatal("expected joined=true")
	}

	members := h.GetMembers("lobby")
	if len(members) != 1 {
		t.Fatalf("expected 1 room member, got %d", len(members))
	}
}

func TestRoomLeaveStep(t *testing.T) {
	h, cleanup := setupTestHub(t)
	defer cleanup()

	h.RegisterConnection("conn-1")
	h.JoinRoom("lobby", "conn-1")

	step, err := newRoomLeaveStep("leave", nil)
	if err != nil {
		t.Fatal(err)
	}

	result, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"connectionId": "conn-1", "room": "lobby"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output["left"] != true {
		t.Fatal("expected left=true")
	}

	members := h.GetMembers("lobby")
	if len(members) != 0 {
		t.Fatalf("expected 0 room members, got %d", len(members))
	}
}

func TestRoomBroadcastStep(t *testing.T) {
	h, cleanup := setupTestHub(t)
	defer cleanup()

	h.RegisterConnection("conn-1")
	h.RegisterConnection("conn-2")
	h.JoinRoom("game-1", "conn-1")
	h.JoinRoom("game-1", "conn-2")

	sent := 0
	h.SetSendFunc(func(connID string, msg []byte) bool {
		sent++
		return true
	})

	step, err := newRoomBroadcastStep("bcast", nil)
	if err != nil {
		t.Fatal(err)
	}

	result, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"room": "game-1", "message": `{"event":"start"}`})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output["recipients"].(int) != 2 {
		t.Fatalf("expected 2 recipients, got %v", result.Output["recipients"])
	}
}

func TestRoomMembersStep(t *testing.T) {
	h, cleanup := setupTestHub(t)
	defer cleanup()

	h.RegisterConnection("conn-1")
	h.RegisterConnection("conn-2")
	h.JoinRoom("room-a", "conn-1")
	h.JoinRoom("room-a", "conn-2")

	step, err := newRoomMembersStep("members", nil)
	if err != nil {
		t.Fatal(err)
	}

	result, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"room": "room-a"})
	if err != nil {
		t.Fatal(err)
	}

	connections := result.Output["connections"].([]string)
	if len(connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(connections))
	}
}

func TestRoomJoinStep_FallbackCurrent(t *testing.T) {
	h, cleanup := setupTestHub(t)
	defer cleanup()

	h.RegisterConnection("conn-x")

	step, err := newRoomJoinStep("join", nil)
	if err != nil {
		t.Fatal(err)
	}

	// connID from current instead of config
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{"connID": "conn-x"}, nil,
		map[string]any{"room": "lobby"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output["joined"] != true {
		t.Fatal("expected joined=true from current fallback")
	}
}
