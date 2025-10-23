package opamp

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/open-telemetry/opamp-go/protobufs"
	"go.uber.org/zap"

	"github.com/cc1024201/opamp-platform/internal/model"
)

func TestUpdateAgentState_NewAgent(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)
	ctx := context.Background()
	agentID := uuid.New().String()
	conn := newMockConnection("conn-1")

	agentUUID := uuid.MustParse(agentID)
	message := &protobufs.AgentToServer{
		InstanceUid: agentUUID[:],
		SequenceNum: 1,
		AgentDescription: &protobufs.AgentDescription{
			IdentifyingAttributes: []*protobufs.KeyValue{
				{
					Key: "service.name",
					Value: &protobufs.AnyValue{
						Value: &protobufs.AnyValue_StringValue{
							StringValue: "test-collector",
						},
					},
				},
				{
					Key: "service.version",
					Value: &protobufs.AnyValue{
						Value: &protobufs.AnyValue_StringValue{
							StringValue: "1.0.0",
						},
					},
				},
				{
					Key: "host.name",
					Value: &protobufs.AnyValue{
						Value: &protobufs.AnyValue_StringValue{
							StringValue: "test-host",
						},
					},
				},
				{
					Key: "host.arch",
					Value: &protobufs.AnyValue{
						Value: &protobufs.AnyValue_StringValue{
							StringValue: "amd64",
						},
					},
				},
				{
					Key: "os.type",
					Value: &protobufs.AnyValue{
						Value: &protobufs.AnyValue_StringValue{
							StringValue: "linux",
						},
					},
				},
			},
			NonIdentifyingAttributes: []*protobufs.KeyValue{
				{
					Key: "env",
					Value: &protobufs.AnyValue{
						Value: &protobufs.AnyValue_StringValue{
							StringValue: "production",
						},
					},
				},
				{
					Key: "region",
					Value: &protobufs.AnyValue{
						Value: &protobufs.AnyValue_StringValue{
							StringValue: "us-west",
						},
					},
				},
			},
		},
	}

	err = opampSrv.updateAgentState(ctx, conn, agentID, message)
	if err != nil {
		t.Fatalf("updateAgentState() failed: %v", err)
	}

	// Verify agent was created
	agent, err := store.GetAgent(ctx, agentID)
	if err != nil {
		t.Fatalf("GetAgent() failed: %v", err)
	}

	if agent == nil {
		t.Fatal("Expected agent to be created")
	}

	// Verify identifying attributes
	if agent.Name != "test-collector" {
		t.Errorf("Name = %v, want test-collector", agent.Name)
	}
	if agent.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", agent.Version)
	}
	if agent.Hostname != "test-host" {
		t.Errorf("Hostname = %v, want test-host", agent.Hostname)
	}
	if agent.Architecture != "amd64" {
		t.Errorf("Architecture = %v, want amd64", agent.Architecture)
	}
	if agent.Type != "linux" {
		t.Errorf("Type = %v, want linux", agent.Type)
	}

	// Verify non-identifying attributes (labels)
	if agent.Labels["env"] != "production" {
		t.Errorf("Label env = %v, want production", agent.Labels["env"])
	}
	if agent.Labels["region"] != "us-west" {
		t.Errorf("Label region = %v, want us-west", agent.Labels["region"])
	}

	// Verify status
	if agent.Status != model.StatusOnline {
		t.Errorf("Status = %v, want %v", agent.Status, model.StatusOnline)
	}

	// Verify sequence number
	if agent.SequenceNumber != 1 {
		t.Errorf("SequenceNumber = %v, want 1", agent.SequenceNumber)
	}

	// Verify connection was registered
	if !opampSrv.connections.isConnected(agentID) {
		t.Error("Expected agent to be registered in connection manager")
	}
}

func TestUpdateAgentState_ExistingAgent(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)
	ctx := context.Background()
	agentID := uuid.New().String()
	conn := newMockConnection("conn-1")

	// Pre-populate store with existing agent
	existingAgent := &model.Agent{
		ID:       agentID,
		Name:     "old-name",
		Version:  "0.5.0",
		Status:   model.StatusOffline,
		Protocol: "opamp",
		Labels:   model.Labels{"old-label": "old-value"},
	}
	store.UpsertAgent(ctx, existingAgent)

	agentUUID := uuid.MustParse(agentID)
	message := &protobufs.AgentToServer{
		InstanceUid: agentUUID[:],
		SequenceNum: 5,
		AgentDescription: &protobufs.AgentDescription{
			IdentifyingAttributes: []*protobufs.KeyValue{
				{
					Key: "service.name",
					Value: &protobufs.AnyValue{
						Value: &protobufs.AnyValue_StringValue{
							StringValue: "updated-collector",
						},
					},
				},
				{
					Key: "service.version",
					Value: &protobufs.AnyValue{
						Value: &protobufs.AnyValue_StringValue{
							StringValue: "2.0.0",
						},
					},
				},
			},
			NonIdentifyingAttributes: []*protobufs.KeyValue{
				{
					Key: "new-label",
					Value: &protobufs.AnyValue{
						Value: &protobufs.AnyValue_StringValue{
							StringValue: "new-value",
						},
					},
				},
			},
		},
	}

	err = opampSrv.updateAgentState(ctx, conn, agentID, message)
	if err != nil {
		t.Fatalf("updateAgentState() failed: %v", err)
	}

	// Verify agent was updated
	agent, err := store.GetAgent(ctx, agentID)
	if err != nil {
		t.Fatalf("GetAgent() failed: %v", err)
	}

	// Verify updates
	if agent.Name != "updated-collector" {
		t.Errorf("Name = %v, want updated-collector", agent.Name)
	}
	if agent.Version != "2.0.0" {
		t.Errorf("Version = %v, want 2.0.0", agent.Version)
	}
	if agent.Status != model.StatusOnline {
		t.Errorf("Status = %v, want %v", agent.Status, model.StatusOnline)
	}
	if agent.SequenceNumber != 5 {
		t.Errorf("SequenceNumber = %v, want 5", agent.SequenceNumber)
	}

	// Verify labels were replaced
	if agent.Labels["new-label"] != "new-value" {
		t.Errorf("Label new-label = %v, want new-value", agent.Labels["new-label"])
	}
}

func TestUpdateAgentState_ConfigFailure(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)
	ctx := context.Background()
	agentID := uuid.New().String()
	conn := newMockConnection("conn-1")

	agentUUID := uuid.MustParse(agentID)
	message := &protobufs.AgentToServer{
		InstanceUid: agentUUID[:],
		SequenceNum: 1,
		RemoteConfigStatus: &protobufs.RemoteConfigStatus{
			Status: protobufs.RemoteConfigStatuses_RemoteConfigStatuses_FAILED,
		},
	}

	err = opampSrv.updateAgentState(ctx, conn, agentID, message)
	if err != nil {
		t.Fatalf("updateAgentState() failed: %v", err)
	}

	// Verify agent status is set to error
	agent, err := store.GetAgent(ctx, agentID)
	if err != nil {
		t.Fatalf("GetAgent() failed: %v", err)
	}

	if agent.Status != model.StatusError {
		t.Errorf("Status = %v, want %v", agent.Status, model.StatusError)
	}
}

func TestCheckAndSendConfig_NoConfig(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)
	ctx := context.Background()
	agentID := uuid.New().String()

	agentUUID := uuid.MustParse(agentID)
	message := &protobufs.AgentToServer{
		InstanceUid: agentUUID[:],
		SequenceNum: 1,
	}

	// No configuration available for agent
	response := opampSrv.checkAndSendConfig(ctx, agentID, message)

	if response != nil {
		t.Error("Expected nil response when no configuration is available")
	}
}

func TestCheckAndSendConfig_NewConfig(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)
	ctx := context.Background()
	agentID := uuid.New().String()

	// Add configuration for agent
	configuration := &model.Configuration{
		Name:       "test-config",
		RawConfig:  "receivers:\n  otlp:",
		ConfigHash: "abc123hash",
	}
	store.configurations[agentID] = configuration

	agentUUID := uuid.MustParse(agentID)
	message := &protobufs.AgentToServer{
		InstanceUid: agentUUID[:],
		SequenceNum: 1,
		RemoteConfigStatus: &protobufs.RemoteConfigStatus{
			LastRemoteConfigHash: []byte("old-hash"),
		},
	}

	response := opampSrv.checkAndSendConfig(ctx, agentID, message)

	// Should send new configuration
	if response == nil {
		t.Fatal("Expected response with new configuration")
	}

	if response.RemoteConfig == nil {
		t.Fatal("Expected RemoteConfig in response")
	}

	if response.RemoteConfig.ConfigHash == nil {
		t.Fatal("Expected ConfigHash in response")
	}

	if string(response.RemoteConfig.ConfigHash) != "abc123hash" {
		t.Errorf("ConfigHash = %v, want abc123hash", string(response.RemoteConfig.ConfigHash))
	}

	if response.RemoteConfig.Config == nil {
		t.Fatal("Expected Config in response")
	}

	configFile := response.RemoteConfig.Config.ConfigMap["config.yaml"]
	if configFile == nil {
		t.Fatal("Expected config.yaml in ConfigMap")
	}

	if string(configFile.Body) != "receivers:\n  otlp:" {
		t.Errorf("Config body = %v, want receivers:\\n  otlp:", string(configFile.Body))
	}
}

func TestCheckAndSendConfig_SameConfig(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)
	ctx := context.Background()
	agentID := uuid.New().String()

	// Add configuration for agent
	configHash := "abc123hash"
	configuration := &model.Configuration{
		Name:       "test-config",
		RawConfig:  "receivers:\n  otlp:",
		ConfigHash: configHash,
	}
	store.configurations[agentID] = configuration

	agentUUID := uuid.MustParse(agentID)
	message := &protobufs.AgentToServer{
		InstanceUid: agentUUID[:],
		SequenceNum: 1,
		RemoteConfigStatus: &protobufs.RemoteConfigStatus{
			LastRemoteConfigHash: []byte(configHash), // Same hash
		},
	}

	response := opampSrv.checkAndSendConfig(ctx, agentID, message)

	// Should NOT send config if hash is the same
	if response != nil {
		t.Error("Expected nil response when config hash matches")
	}
}

func TestOnConnectionClose(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)
	ctx := context.Background()
	agentID := uuid.New().String()
	conn := newMockConnection("conn-1")

	// Create and register agent
	agent := &model.Agent{
		ID:       agentID,
		Name:     "test-agent",
		Status:   model.StatusOnline,
		Protocol: "opamp",
		Labels:   make(model.Labels),
	}
	store.UpsertAgent(ctx, agent)
	opampSrv.connections.addConnection(agentID, conn)

	// Simulate connection close
	opampSrv.onConnectionClose(conn)

	// Verify connection was removed
	if opampSrv.connections.isConnected(agentID) {
		t.Error("Expected agent to be disconnected")
	}

	// Verify agent status was updated to disconnected
	updatedAgent, err := store.GetAgent(ctx, agentID)
	if err != nil {
		t.Fatalf("GetAgent() failed: %v", err)
	}

	if updatedAgent.Status != model.StatusOffline {
		t.Errorf("Status = %v, want %v", updatedAgent.Status, model.StatusOffline)
	}

	if updatedAgent.LastDisconnectedAt == nil {
		t.Error("Expected LastDisconnectedAt to be set")
	}
}

func TestOnConnectionClose_NonExistentAgent(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)
	conn := newMockConnection("conn-unknown")

	// Connection not in manager - should not panic
	opampSrv.onConnectionClose(conn)
}
