package internal

import (
	"fmt"
	"sync"

	"github.com/GoCodeAlone/workflow-plugin-rooms/internal/contracts"
	pb "github.com/GoCodeAlone/workflow/plugin/external/proto"
	"github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
)

// Version is set at build time via -ldflags
// "-X github.com/GoCodeAlone/workflow-plugin-rooms/internal.Version=X.Y.Z"
var Version = "dev"

var (
	globalHub   *Hub
	globalHubMu sync.RWMutex
)

// GetGlobalHub returns the global room hub once the rooms.manager module has initialized.
func GetGlobalHub() *Hub {
	globalHubMu.RLock()
	defer globalHubMu.RUnlock()
	return globalHub
}

// SetGlobalHub sets the global hub (called by the module on Init).
func SetGlobalHub(h *Hub) {
	globalHubMu.Lock()
	globalHub = h
	globalHubMu.Unlock()
}

type roomsPlugin struct{}

// NewRoomsPlugin returns the rooms SDK plugin provider.
func NewRoomsPlugin() sdk.PluginProvider {
	return &roomsPlugin{}
}

func (p *roomsPlugin) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "workflow-plugin-rooms",
		Version:     Version,
		Author:      "GoCodeAlone",
		Description: "Room management for workflow applications: join, leave, broadcast, members",
	}
}

var roomsModuleTypes = []string{"rooms.manager"}

func (p *roomsPlugin) ModuleTypes() []string {
	return append([]string(nil), roomsModuleTypes...)
}

var roomsStepTypes = []string{
	"step.room_join",
	"step.room_leave",
	"step.room_broadcast",
	"step.room_members",
}

func (p *roomsPlugin) StepTypes() []string {
	return append([]string(nil), roomsStepTypes...)
}

func (p *roomsPlugin) CreateModule(typeName, name string, config map[string]any) (sdk.ModuleInstance, error) {
	switch typeName {
	case "rooms.manager":
		return newRoomsManagerModule(name, config)
	default:
		return nil, fmt.Errorf("unknown module type %q", typeName)
	}
}

// TypedModuleTypes returns the protobuf-typed module type names this plugin provides.
func (p *roomsPlugin) TypedModuleTypes() []string {
	return append([]string(nil), roomsModuleTypes...)
}

// CreateTypedModule creates a typed module instance of the given type.
func (p *roomsPlugin) CreateTypedModule(typeName, name string, config *anypb.Any) (sdk.ModuleInstance, error) {
	switch typeName {
	case "rooms.manager":
		factory := sdk.NewTypedModuleFactory(typeName, &contracts.RoomsManagerConfig{}, func(name string, _ *contracts.RoomsManagerConfig) (sdk.ModuleInstance, error) {
			return newRoomsManagerModule(name, nil)
		})
		return factory.CreateTypedModule(typeName, name, config)
	default:
		return nil, fmt.Errorf("unknown typed module type %q", typeName)
	}
}

// TypedStepTypes returns the protobuf-typed step type names this plugin provides.
func (p *roomsPlugin) TypedStepTypes() []string {
	return append([]string(nil), roomsStepTypes...)
}

// CreateTypedStep creates a typed step instance of the given type.
func (p *roomsPlugin) CreateTypedStep(typeName, name string, config *anypb.Any) (sdk.StepInstance, error) {
	switch typeName {
	case "step.room_join":
		factory := sdk.NewTypedStepFactory(typeName, &contracts.RoomJoinConfig{}, &contracts.RoomJoinInput{}, typedRoomJoin)
		return factory.CreateTypedStep(typeName, name, config)
	case "step.room_leave":
		factory := sdk.NewTypedStepFactory(typeName, &contracts.RoomLeaveConfig{}, &contracts.RoomLeaveInput{}, typedRoomLeave)
		return factory.CreateTypedStep(typeName, name, config)
	case "step.room_broadcast":
		factory := sdk.NewTypedStepFactory(typeName, &contracts.RoomBroadcastConfig{}, &contracts.RoomBroadcastInput{}, typedRoomBroadcast)
		return factory.CreateTypedStep(typeName, name, config)
	case "step.room_members":
		factory := sdk.NewTypedStepFactory(typeName, &contracts.RoomMembersConfig{}, &contracts.RoomMembersInput{}, typedRoomMembers)
		return factory.CreateTypedStep(typeName, name, config)
	default:
		return nil, fmt.Errorf("unknown typed step type %q", typeName)
	}
}

// ContractRegistry returns strict protobuf descriptors for plugin boundaries.
func (p *roomsPlugin) ContractRegistry() *pb.ContractRegistry {
	const pkg = "workflow.plugins.rooms.v1."
	return &pb.ContractRegistry{
		FileDescriptorSet: &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{
				protodesc.ToFileDescriptorProto(contracts.File_internal_contracts_rooms_proto),
			},
		},
		Contracts: []*pb.ContractDescriptor{
			moduleContract("rooms.manager", pkg+"RoomsManagerConfig"),
			stepContract("step.room_join", pkg+"RoomJoinConfig", pkg+"RoomJoinInput", pkg+"RoomJoinOutput"),
			stepContract("step.room_leave", pkg+"RoomLeaveConfig", pkg+"RoomLeaveInput", pkg+"RoomLeaveOutput"),
			stepContract("step.room_broadcast", pkg+"RoomBroadcastConfig", pkg+"RoomBroadcastInput", pkg+"RoomBroadcastOutput"),
			stepContract("step.room_members", pkg+"RoomMembersConfig", pkg+"RoomMembersInput", pkg+"RoomMembersOutput"),
		},
	}
}

func moduleContract(moduleType, configMessage string) *pb.ContractDescriptor {
	return &pb.ContractDescriptor{
		Kind:          pb.ContractKind_CONTRACT_KIND_MODULE,
		ModuleType:    moduleType,
		ConfigMessage: configMessage,
		Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
	}
}

func stepContract(stepType, configMessage, inputMessage, outputMessage string) *pb.ContractDescriptor {
	return &pb.ContractDescriptor{
		Kind:          pb.ContractKind_CONTRACT_KIND_STEP,
		StepType:      stepType,
		ConfigMessage: configMessage,
		InputMessage:  inputMessage,
		OutputMessage: outputMessage,
		Mode:          pb.ContractMode_CONTRACT_MODE_STRICT_PROTO,
	}
}

func (p *roomsPlugin) CreateStep(typeName, name string, config map[string]any) (sdk.StepInstance, error) {
	switch typeName {
	case "step.room_join":
		return newRoomJoinStep(name, config)
	case "step.room_leave":
		return newRoomLeaveStep(name, config)
	case "step.room_broadcast":
		return newRoomBroadcastStep(name, config)
	case "step.room_members":
		return newRoomMembersStep(name, config)
	default:
		return nil, fmt.Errorf("unknown step type %q", typeName)
	}
}
