package internal

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-rooms/internal/contracts"
	pb "github.com/GoCodeAlone/workflow/plugin/external/proto"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestRoomsPluginTypedContracts(t *testing.T) {
	provider := NewRoomsPlugin()
	moduleProvider, ok := provider.(sdk.ModuleProvider)
	if !ok {
		t.Fatal("expected module provider")
	}
	typedModuleProvider, ok := provider.(sdk.TypedModuleProvider)
	if !ok {
		t.Fatal("expected typed module provider")
	}
	stepProvider, ok := provider.(sdk.StepProvider)
	if !ok {
		t.Fatal("expected step provider")
	}
	typedStepProvider, ok := provider.(sdk.TypedStepProvider)
	if !ok {
		t.Fatal("expected typed step provider")
	}
	contractProvider, ok := provider.(sdk.ContractProvider)
	if !ok {
		t.Fatal("expected contract provider")
	}

	wantModules := []string{"rooms.manager"}
	wantSteps := []string{
		"step.room_join",
		"step.room_leave",
		"step.room_broadcast",
		"step.room_members",
	}
	assertStringSet(t, typedModuleProvider.TypedModuleTypes(), wantModules)
	assertStringSet(t, typedStepProvider.TypedStepTypes(), wantSteps)
	assertProviderSlicesAreDefensive(t, moduleProvider, typedModuleProvider, stepProvider, typedStepProvider)

	registry := contractProvider.ContractRegistry()
	if registry == nil {
		t.Fatal("expected contract registry")
	}
	if registry.FileDescriptorSet == nil || len(registry.FileDescriptorSet.File) == 0 {
		t.Fatal("expected file descriptor set")
	}
	files, err := protodesc.NewFiles(registry.FileDescriptorSet)
	if err != nil {
		t.Fatalf("descriptor set: %v", err)
	}

	manifestContracts := loadManifestContracts(t)
	contractsByKey := map[string]*pb.ContractDescriptor{}
	for _, descriptor := range registry.Contracts {
		if descriptor.Mode != pb.ContractMode_CONTRACT_MODE_STRICT_PROTO {
			t.Fatalf("%s mode = %s, want strict proto", contractKey(descriptor), descriptor.Mode)
		}
		switch descriptor.Kind {
		case pb.ContractKind_CONTRACT_KIND_MODULE:
			contractsByKey["module:"+descriptor.ModuleType] = descriptor
		case pb.ContractKind_CONTRACT_KIND_STEP:
			contractsByKey["step:"+descriptor.StepType] = descriptor
		default:
			t.Fatalf("unexpected contract kind %s", descriptor.Kind)
		}
		for _, name := range []string{descriptor.ConfigMessage, descriptor.InputMessage, descriptor.OutputMessage} {
			if name == "" {
				continue
			}
			if _, err := files.FindDescriptorByName(protoreflect.FullName(name)); err != nil {
				t.Fatalf("%s references unknown message %s: %v", contractKey(descriptor), name, err)
			}
		}
	}

	for _, key := range []string{
		"module:rooms.manager",
		"step:step.room_join",
		"step:step.room_leave",
		"step:step.room_broadcast",
		"step:step.room_members",
	} {
		contract, ok := contractsByKey[key]
		if !ok {
			t.Fatalf("missing contract %s", key)
		}
		if want, ok := manifestContracts[key]; !ok {
			t.Fatalf("%s missing from plugin.contracts.json", key)
		} else if want.ConfigMessage != contract.ConfigMessage || want.InputMessage != contract.InputMessage || want.OutputMessage != contract.OutputMessage {
			t.Fatalf("%s manifest contract = %#v, runtime = %#v", key, want, contract)
		}
	}
	if len(manifestContracts) != len(contractsByKey) {
		t.Fatalf("plugin.contracts.json contract count = %d, runtime = %d", len(manifestContracts), len(contractsByKey))
	}
}

func TestTypedRoomJoinProviderValidatesTypedConfig(t *testing.T) {
	provider := NewRoomsPlugin().(sdk.TypedStepProvider)
	config, err := anypb.New(&contracts.RoomJoinConfig{Room: "lobby"})
	if err != nil {
		t.Fatalf("pack config: %v", err)
	}
	step, err := provider.CreateTypedStep("step.room_join", "join", config)
	if err != nil {
		t.Fatalf("CreateTypedStep: %v", err)
	}
	if _, err := step.Execute(context.Background(), nil, nil, nil, nil, nil); err == nil {
		t.Fatal("legacy Execute succeeded for typed-only step")
	}

	wrongConfig, err := anypb.New(&contracts.RoomMembersConfig{Room: "lobby"})
	if err != nil {
		t.Fatalf("pack wrong config: %v", err)
	}
	if _, err := provider.CreateTypedStep("step.room_join", "join", wrongConfig); err == nil {
		t.Fatal("CreateTypedStep accepted wrong typed config")
	}
}

func TestTypedRoomJoinMergesConfigInputAndCurrent(t *testing.T) {
	hub := NewHub()
	hub.RegisterConnection("conn-1")
	SetGlobalHub(hub)
	t.Cleanup(func() { SetGlobalHub(nil) })

	result, err := typedRoomJoin(context.Background(), sdk.TypedStepRequest[*contracts.RoomJoinConfig, *contracts.RoomJoinInput]{
		Config:  &contracts.RoomJoinConfig{Room: "lobby"},
		Input:   &contracts.RoomJoinInput{},
		Current: map[string]any{"connectionId": "conn-1"},
	})
	if err != nil {
		t.Fatalf("typedRoomJoin: %v", err)
	}
	if result.Output.GetError() != "" {
		t.Fatalf("error = %q", result.Output.GetError())
	}
	if !result.Output.GetJoined() {
		t.Fatal("joined = false, want true")
	}
	if got := hub.GetMembers("lobby"); len(got) != 1 || got[0] != "conn-1" {
		t.Fatalf("members = %v, want [conn-1]", got)
	}
}

func TestTypedRoomBroadcastRequiresRoomAndMessage(t *testing.T) {
	_, err := typedRoomBroadcast(context.Background(), sdk.TypedStepRequest[*contracts.RoomBroadcastConfig, *contracts.RoomBroadcastInput]{
		Config: &contracts.RoomBroadcastConfig{Message: "hello"},
		Input:  &contracts.RoomBroadcastInput{},
	})
	if err == nil {
		t.Fatal("typedRoomBroadcast accepted missing room")
	}

	_, err = typedRoomBroadcast(context.Background(), sdk.TypedStepRequest[*contracts.RoomBroadcastConfig, *contracts.RoomBroadcastInput]{
		Config: &contracts.RoomBroadcastConfig{Room: "lobby"},
		Input:  &contracts.RoomBroadcastInput{},
	})
	if err == nil {
		t.Fatal("typedRoomBroadcast accepted missing message")
	}
}

func TestTypedRoomMembersRequiresRoom(t *testing.T) {
	_, err := typedRoomMembers(context.Background(), sdk.TypedStepRequest[*contracts.RoomMembersConfig, *contracts.RoomMembersInput]{
		Config: &contracts.RoomMembersConfig{},
		Input:  &contracts.RoomMembersInput{},
	})
	if err == nil {
		t.Fatal("typedRoomMembers accepted missing room")
	}
}

func TestPluginJSONDeclaresEngineAndCapabilities(t *testing.T) {
	var manifest struct {
		Name             string `json:"name"`
		Type             string `json:"type"`
		MinEngineVersion string `json:"minEngineVersion"`
		Capabilities     struct {
			ModuleTypes []string `json:"moduleTypes"`
			StepTypes   []string `json:"stepTypes"`
		} `json:"capabilities"`
	}
	readJSONFile(t, "plugin.json", &manifest)
	if manifest.Name != "workflow-plugin-rooms" {
		t.Fatalf("name = %q, want workflow-plugin-rooms", manifest.Name)
	}
	if manifest.Type != "external" {
		t.Fatalf("type = %q, want external", manifest.Type)
	}
	if manifest.MinEngineVersion != "0.19.0" {
		t.Fatalf("minEngineVersion = %q, want 0.19.0", manifest.MinEngineVersion)
	}
	assertStringSet(t, manifest.Capabilities.ModuleTypes, []string{"rooms.manager"})
	assertStringSet(t, manifest.Capabilities.StepTypes, []string{
		"step.room_join",
		"step.room_leave",
		"step.room_broadcast",
		"step.room_members",
	})
}

type manifestContract struct {
	Mode          string `json:"mode"`
	ConfigMessage string `json:"config"`
	InputMessage  string `json:"input"`
	OutputMessage string `json:"output"`
}

func loadManifestContracts(t *testing.T) map[string]manifestContract {
	t.Helper()
	var manifest struct {
		Version   string `json:"version"`
		Contracts []struct {
			Kind string `json:"kind"`
			Type string `json:"type"`
			manifestContract
		} `json:"contracts"`
	}
	readJSONFile(t, "plugin.contracts.json", &manifest)
	if manifest.Version != "v1" {
		t.Fatalf("plugin.contracts.json version = %q, want v1", manifest.Version)
	}
	contracts := make(map[string]manifestContract, len(manifest.Contracts))
	for _, contract := range manifest.Contracts {
		if contract.Mode != "strict" {
			t.Fatalf("%s mode = %q, want strict", contract.Type, contract.Mode)
		}
		var key string
		switch contract.Kind {
		case "module":
			key = "module:" + contract.Type
		case "step":
			key = "step:" + contract.Type
		default:
			t.Fatalf("unexpected contract kind %q in plugin.contracts.json", contract.Kind)
		}
		if _, exists := contracts[key]; exists {
			t.Fatalf("duplicate contract %q in plugin.contracts.json", key)
		}
		contracts[key] = contract.manifestContract
	}
	return contracts
}

func readJSONFile(t *testing.T, name string, out any) {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
}

func assertProviderSlicesAreDefensive(t *testing.T, moduleProvider sdk.ModuleProvider, typedModuleProvider sdk.TypedModuleProvider, stepProvider sdk.StepProvider, typedStepProvider sdk.TypedStepProvider) {
	t.Helper()
	moduleTypes := moduleProvider.ModuleTypes()
	moduleTypes[0] = "mutated"
	if got := moduleProvider.ModuleTypes()[0]; got == "mutated" {
		t.Fatal("ModuleTypes exposed mutable package-level slice")
	}

	typedModuleTypes := typedModuleProvider.TypedModuleTypes()
	typedModuleTypes[0] = "mutated"
	if got := typedModuleProvider.TypedModuleTypes()[0]; got == "mutated" {
		t.Fatal("TypedModuleTypes exposed mutable package-level slice")
	}

	stepTypes := stepProvider.StepTypes()
	stepTypes[0] = "mutated"
	if got := stepProvider.StepTypes()[0]; got == "mutated" {
		t.Fatal("StepTypes exposed mutable package-level slice")
	}

	typedStepTypes := typedStepProvider.TypedStepTypes()
	typedStepTypes[0] = "mutated"
	if got := typedStepProvider.TypedStepTypes()[0]; got == "mutated" {
		t.Fatal("TypedStepTypes exposed mutable package-level slice")
	}
}

func assertStringSet(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: got %v", len(got), len(want), got)
	}
	seen := make(map[string]bool, len(got))
	for _, item := range got {
		seen[item] = true
	}
	for _, item := range want {
		if !seen[item] {
			t.Fatalf("missing %q in %v", item, got)
		}
	}
}

func contractKey(contract *pb.ContractDescriptor) string {
	switch contract.Kind {
	case pb.ContractKind_CONTRACT_KIND_MODULE:
		return "module:" + contract.ModuleType
	case pb.ContractKind_CONTRACT_KIND_STEP:
		return "step:" + contract.StepType
	default:
		return contract.Kind.String()
	}
}
