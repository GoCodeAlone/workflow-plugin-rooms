# workflow-plugin-rooms

Room management plugin for the [workflow](https://github.com/GoCodeAlone/workflow) framework. Extracted from `workflow-plugin-websocket` to provide a standalone, thread-safe room hub for connection grouping, broadcast, and membership queries.

## Features

- Join/leave named rooms by connection ID
- Broadcast messages to a room with optional sender exclusion
- Query room membership
- Register/unregister connections with automatic room cleanup
- Empty rooms are garbage-collected on leave/unregister

## RoomHub Interface

```go
type RoomHub interface {
    JoinRoom(room, connID string) bool
    LeaveRoom(room, connID string)
    LeaveAllRooms(connID string)
    BroadcastToRoom(room string, msg []byte, excludeConnID string) int
    GetMembers(room string) []string
    RegisterConnection(connID string)
    UnregisterConnection(connID string)
}
```

## Pipeline Steps

| Step Type | Description | Config Params |
|---|---|---|
| `step.room_join` | Add connection to a room | `room`, `connectionId` |
| `step.room_leave` | Remove connection from a room | `room`, `connectionId` |
| `step.room_broadcast` | Send message to all room members | `room`, `message`, `exclude` |
| `step.room_members` | Query connections in a room | `room` |

Module type: `rooms.manager`

## Usage

```go
import rooms "github.com/GoCodeAlone/workflow-plugin-rooms"

provider := rooms.NewRoomsPlugin()
// register with workflow engine

hub := rooms.GetHub() // after module init
hub.JoinRoom("lobby", connID)
```

## Build & Test

```sh
GOWORK=off go build ./...
GOWORK=off go test ./...
```

Requires Go 1.26+ and `github.com/GoCodeAlone/workflow` v0.19.0.

## License

MIT
