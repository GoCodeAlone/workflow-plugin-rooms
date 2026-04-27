// Command workflow-plugin-rooms is a workflow engine external plugin that
// provides room management module and step types.
package main

import (
	"github.com/GoCodeAlone/workflow-plugin-rooms/internal"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func main() {
	sdk.Serve(internal.NewRoomsPlugin())
}
