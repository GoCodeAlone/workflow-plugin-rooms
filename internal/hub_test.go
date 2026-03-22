package internal

import (
	"testing"
)

func TestHub_JoinLeave(t *testing.T) {
	h := NewHub()

	h.RegisterConnection("conn-1")

	joined := h.JoinRoom("game-room-1", "conn-1")
	if !joined {
		t.Fatal("expected join to succeed")
	}

	members := h.GetMembers("game-room-1")
	if len(members) != 1 || members[0] != "conn-1" {
		t.Fatalf("expected [conn-1], got %v", members)
	}

	h.LeaveRoom("game-room-1", "conn-1")

	members = h.GetMembers("game-room-1")
	if len(members) != 0 {
		t.Fatalf("expected empty room, got %v", members)
	}
}

func TestHub_JoinRequiresRegistration(t *testing.T) {
	h := NewHub()

	// Should fail — connection not registered
	joined := h.JoinRoom("room-a", "unknown-conn")
	if joined {
		t.Fatal("expected join to fail for unregistered connection")
	}
}

func TestHub_LeaveAllRooms(t *testing.T) {
	h := NewHub()
	h.RegisterConnection("conn-1")

	h.JoinRoom("room-a", "conn-1")
	h.JoinRoom("room-b", "conn-1")

	h.LeaveAllRooms("conn-1")

	if len(h.GetMembers("room-a")) != 0 {
		t.Fatal("conn-1 should have left room-a")
	}
	if len(h.GetMembers("room-b")) != 0 {
		t.Fatal("conn-1 should have left room-b")
	}
}

func TestHub_UnregisterConnection(t *testing.T) {
	h := NewHub()
	h.RegisterConnection("conn-1")
	h.RegisterConnection("conn-2")
	h.JoinRoom("room-a", "conn-1")
	h.JoinRoom("room-a", "conn-2")

	h.UnregisterConnection("conn-1")

	members := h.GetMembers("room-a")
	if len(members) != 1 || members[0] != "conn-2" {
		t.Fatalf("expected [conn-2], got %v", members)
	}
}

func TestHub_BroadcastToRoom(t *testing.T) {
	h := NewHub()
	h.RegisterConnection("conn-1")
	h.RegisterConnection("conn-2")
	h.JoinRoom("room-a", "conn-1")
	h.JoinRoom("room-a", "conn-2")

	sent := make(map[string][]byte)
	h.SetSendFunc(func(connID string, msg []byte) bool {
		cp := make([]byte, len(msg))
		copy(cp, msg)
		sent[connID] = cp
		return true
	})

	count := h.BroadcastToRoom("room-a", []byte(`{"event":"test"}`), "")
	if count != 2 {
		t.Fatalf("expected 2 recipients, got %d", count)
	}
	if string(sent["conn-1"]) != `{"event":"test"}` || string(sent["conn-2"]) != `{"event":"test"}` {
		t.Fatal("both connections should receive the message")
	}
}

func TestHub_BroadcastExclude(t *testing.T) {
	h := NewHub()
	h.RegisterConnection("conn-1")
	h.RegisterConnection("conn-2")
	h.JoinRoom("room-a", "conn-1")
	h.JoinRoom("room-a", "conn-2")

	sent := make(map[string]bool)
	h.SetSendFunc(func(connID string, msg []byte) bool {
		sent[connID] = true
		return true
	})

	count := h.BroadcastToRoom("room-a", []byte(`test`), "conn-1")
	if count != 1 {
		t.Fatalf("expected 1 recipient, got %d", count)
	}
	if sent["conn-1"] {
		t.Fatal("conn-1 should be excluded")
	}
	if !sent["conn-2"] {
		t.Fatal("conn-2 should receive broadcast")
	}
}

func TestHub_BroadcastNoSendFunc(t *testing.T) {
	h := NewHub()
	h.RegisterConnection("conn-1")
	h.JoinRoom("room-a", "conn-1")

	// No sendFunc set — should return 0 without panic
	count := h.BroadcastToRoom("room-a", []byte(`test`), "")
	if count != 0 {
		t.Fatalf("expected 0 recipients without sendFunc, got %d", count)
	}
}

func TestHub_GetMembers_EmptyRoom(t *testing.T) {
	h := NewHub()
	members := h.GetMembers("nonexistent")
	if members != nil {
		t.Fatalf("expected nil for nonexistent room, got %v", members)
	}
}
