package internal

import (
	"sync"
)

// RoomHub is the public interface for room management.
type RoomHub interface {
	JoinRoom(room, connID string) bool
	LeaveRoom(room, connID string)
	LeaveAllRooms(connID string)
	BroadcastToRoom(room string, msg []byte, excludeConnID string) int
	GetMembers(room string) []string
	RegisterConnection(connID string)
	UnregisterConnection(connID string)
}

// Hub implements RoomHub with a concurrent-safe room registry.
type Hub struct {
	rooms     map[string]map[string]bool // room -> set of connIDs
	connRooms map[string]map[string]bool // connID -> set of rooms
	mu        sync.RWMutex

	// sendFunc is the hook used to deliver messages. Must be set by the host
	// before broadcasting. If nil, BroadcastToRoom is a no-op.
	sendFunc func(connID string, msg []byte) bool
}

// NewHub creates a new room hub.
func NewHub() *Hub {
	return &Hub{
		rooms:     make(map[string]map[string]bool),
		connRooms: make(map[string]map[string]bool),
	}
}

// SetSendFunc installs the function used by BroadcastToRoom to deliver messages.
func (h *Hub) SetSendFunc(f func(connID string, msg []byte) bool) {
	h.mu.Lock()
	h.sendFunc = f
	h.mu.Unlock()
}

// RegisterConnection registers a connection ID so it can join rooms.
func (h *Hub) RegisterConnection(connID string) {
	h.mu.Lock()
	if _, ok := h.connRooms[connID]; !ok {
		h.connRooms[connID] = make(map[string]bool)
	}
	h.mu.Unlock()
}

// UnregisterConnection removes a connection from all rooms and cleans up.
func (h *Hub) UnregisterConnection(connID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for room := range h.connRooms[connID] {
		delete(h.rooms[room], connID)
		if len(h.rooms[room]) == 0 {
			delete(h.rooms, room)
		}
	}
	delete(h.connRooms, connID)
}

// JoinRoom adds connID to the named room. Returns true if the connection was registered.
func (h *Hub) JoinRoom(room, connID string) bool {
	if connID == "" || room == "" {
		return false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.connRooms[connID]; !ok {
		return false
	}
	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[string]bool)
	}
	h.rooms[room][connID] = true
	h.connRooms[connID][room] = true
	return true
}

// LeaveRoom removes connID from the named room.
func (h *Hub) LeaveRoom(room, connID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if members, ok := h.rooms[room]; ok {
		delete(members, connID)
		if len(members) == 0 {
			delete(h.rooms, room)
		}
	}
	if rooms, ok := h.connRooms[connID]; ok {
		delete(rooms, room)
	}
}

// LeaveAllRooms removes connID from every room it belongs to.
func (h *Hub) LeaveAllRooms(connID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for room := range h.connRooms[connID] {
		delete(h.rooms[room], connID)
		if len(h.rooms[room]) == 0 {
			delete(h.rooms, room)
		}
	}
	delete(h.connRooms, connID)
}

// BroadcastToRoom sends msg to all connections in room (optionally excluding one).
// Returns the number of recipients that accepted the message.
func (h *Hub) BroadcastToRoom(room string, msg []byte, excludeConnID string) int {
	h.mu.RLock()
	members, ok := h.rooms[room]
	if !ok {
		h.mu.RUnlock()
		return 0
	}
	ids := make([]string, 0, len(members))
	for id := range members {
		if id != excludeConnID {
			ids = append(ids, id)
		}
	}
	sendFn := h.sendFunc
	h.mu.RUnlock()

	if sendFn == nil {
		return 0
	}
	count := 0
	for _, id := range ids {
		if sendFn(id, msg) {
			count++
		}
	}
	return count
}

// GetMembers returns a list of connection IDs in the named room.
func (h *Hub) GetMembers(room string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	members, ok := h.rooms[room]
	if !ok {
		return nil
	}
	result := make([]string, 0, len(members))
	for id := range members {
		result = append(result, id)
	}
	return result
}
