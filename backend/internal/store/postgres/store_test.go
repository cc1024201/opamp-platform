package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cc1024201/opamp-platform/internal/model"
	"go.uber.org/zap"
)

var testStore *Store

func TestMain(m *testing.M) {
	// 设置测试数据库
	logger, _ := zap.NewDevelopment()

	// 使用环境变量或默认测试数据库配置
	config := Config{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefaultInt("TEST_DB_PORT", 5432),
		User:     getEnvOrDefault("TEST_DB_USER", "opamp"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "opamp123"),
		DBName:   getEnvOrDefault("TEST_DB_NAME", "opamp_platform"),
		SSLMode:  "disable",
	}

	var err error
	testStore, err = NewStore(config, logger)
	if err != nil {
		logger.Fatal("Failed to create test store", zap.Error(err))
	}

	// 运行测试
	code := m.Run()

	// 清理
	if testStore != nil {
		testStore.Close()
	}

	os.Exit(code)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func cleanupDatabase(t *testing.T) {
	testStore.db.Exec("TRUNCATE TABLE agents, configurations, sources, destinations, processors, users CASCADE")
	t.Cleanup(func() {
		testStore.db.Exec("TRUNCATE TABLE agents, configurations, sources, destinations, processors, users CASCADE")
	})
}

func TestStore_UpsertAgent(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	agent := &model.Agent{
		ID:       "test-agent-001",
		Name:     "test-collector",
		Type:     "otel-collector",
		Hostname: "test-host-001",
		Version:  "1.0.0",
		Status:   model.StatusOnline,
		Labels: model.Labels{
			"env":    "test",
			"region": "us-west",
		},
		Protocol:       "opamp",
		SequenceNumber: 1,
	}

	// Test create
	err := testStore.UpsertAgent(ctx, agent)
	if err != nil {
		t.Fatalf("UpsertAgent (create) failed: %v", err)
	}

	// Verify creation
	retrieved, err := testStore.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("GetAgent failed: %v", err)
	}
	if retrieved.ID != agent.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, agent.ID)
	}
	if retrieved.Labels["env"] != "test" {
		t.Errorf("Label env = %v, want test", retrieved.Labels["env"])
	}

	// Test update
	agent.Version = "2.0.0"
	agent.Labels["env"] = "production"
	err = testStore.UpsertAgent(ctx, agent)
	if err != nil {
		t.Fatalf("UpsertAgent (update) failed: %v", err)
	}

	// Verify update
	retrieved, err = testStore.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("GetAgent failed after update: %v", err)
	}
	if retrieved.Version != "2.0.0" {
		t.Errorf("Version = %v, want 2.0.0", retrieved.Version)
	}
	if retrieved.Labels["env"] != "production" {
		t.Errorf("Label env = %v, want production", retrieved.Labels["env"])
	}
}

func TestStore_GetAgent(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Test getting non-existent agent
	agent, err := testStore.GetAgent(ctx, "non-existent-id")
	if err != nil || agent != nil {
		t.Error("Expected nil agent and nil error for non-existent agent")
	}

	// Create and get agent
	newAgent := &model.Agent{
		ID:       "test-newAgent-002",
		Name:     "test-collector-2",
		Hostname: "test-host-002",
		Status:   model.StatusOnline,
		Labels: model.Labels{
			"os.type": "linux",
		},
		Protocol: "opamp",
	}

	if err := testStore.UpsertAgent(ctx, newAgent); err != nil {
		t.Fatalf("UpsertAgent failed: %v", err)
	}

	retrieved, err := testStore.GetAgent(ctx, newAgent.ID)
	if err != nil {
		t.Fatalf("GetAgent failed: %v", err)
	}
	if retrieved.ID != newAgent.ID {
		t.Error("Retrieved newAgent ID mismatch")
	}
}

func TestStore_ListAgents(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Create multiple agents
	for i := 1; i <= 3; i++ {
		agent := &model.Agent{
			ID:       fmt.Sprintf("list-agent-%03d", i),
			Name:     fmt.Sprintf("collector-%d", i),
			Hostname: fmt.Sprintf("host-%03d", i),
			Status:   model.StatusOnline,
			Protocol: "opamp",
		}
		if err := testStore.UpsertAgent(ctx, agent); err != nil {
			t.Fatalf("UpsertAgent failed: %v", err)
		}
	}

	// Test list
	list, total, err := testStore.ListAgents(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListAgents failed: %v", err)
	}
	if total != 3 {
		t.Errorf("Total = %v, want 3", total)
	}
	if len(list) != 3 {
		t.Errorf("List length = %v, want 3", len(list))
	}

	// Test pagination
	list, _, err = testStore.ListAgents(ctx, 2, 0)
	if err != nil {
		t.Fatalf("ListAgents with limit failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List length with limit = %v, want 2", len(list))
	}
}

func TestStore_DeleteAgent(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	agent := &model.Agent{
		ID:       "test-agent-del",
		Name:     "delete-test",
		Hostname: "test-host",
		Status:   model.StatusOnline,
		Protocol: "opamp",
	}

	if err := testStore.UpsertAgent(ctx, agent); err != nil {
		t.Fatalf("UpsertAgent failed: %v", err)
	}

	// Delete
	if err := testStore.DeleteAgent(ctx, agent.ID); err != nil {
		t.Fatalf("DeleteAgent failed: %v", err)
	}

	// Verify deletion
	deletedAgent, err := testStore.GetAgent(ctx, agent.ID)
	if err != nil || deletedAgent != nil {
		t.Errorf("Expected nil agent and nil error for deleted agent, got agent=%v, err=%v", 
			deletedAgent, err)
	}
}

func TestStore_CreateConfiguration(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	config := &model.Configuration{
		Name:        "test-config-001",
		DisplayName: "Test Config",
		ContentType: "yaml",
		RawConfig:   "receivers:\n  otlp:",
		Selector: map[string]string{
			"env": "test",
		},
	}
	config.UpdateHash()

	err := testStore.CreateConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("CreateConfiguration failed: %v", err)
	}

	// Verify
	retrieved, err := testStore.GetConfigurationByName(ctx, config.Name)
	if err != nil {
		t.Fatalf("GetConfigurationByName failed: %v", err)
	}
	if retrieved.Name != config.Name {
		t.Errorf("Name = %v, want %v", retrieved.Name, config.Name)
	}
	if retrieved.Selector["env"] != "test" {
		t.Errorf("Selector env = %v, want test", retrieved.Selector["env"])
	}
}

func TestStore_UpdateConfiguration(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	config := &model.Configuration{
		Name:        "update-config",
		DisplayName: "Original",
		ContentType: "yaml",
		RawConfig:   "receivers:\n  otlp:",
	}
	config.UpdateHash()

	if err := testStore.CreateConfiguration(ctx, config); err != nil {
		t.Fatalf("CreateConfiguration failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	// Update
	config.DisplayName = "Updated"
	config.RawConfig = "receivers:\n  otlp:\n    protocols:\n      grpc:"
	config.UpdateHash()

	if err := testStore.UpdateConfiguration(ctx, config); err != nil {
		t.Fatalf("UpdateConfiguration failed: %v", err)
	}

	// Verify
	retrieved, err := testStore.GetConfigurationByName(ctx, config.Name)
	if err != nil {
		t.Fatalf("GetConfigurationByName failed: %v", err)
	}
	if retrieved.DisplayName != "Updated" {
		t.Errorf("DisplayName = %v, want Updated", retrieved.DisplayName)
	}
}

func TestStore_GetConfiguration(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Create agent
	agent := &model.Agent{
		ID:       "config-test-agent",
		Name:     "test-collector",
		Hostname: "test-host",
		Status:   model.StatusOnline,
		Labels: model.Labels{
			"os.type": "linux",
		},
		Protocol: "opamp",
	}
	if err := testStore.UpsertAgent(ctx, agent); err != nil {
		t.Fatalf("UpsertAgent failed: %v", err)
	}

	// Create matching configuration
	config := &model.Configuration{
		Name:        "linux-config",
		DisplayName: "Linux Config",
		ContentType: "yaml",
		RawConfig:   "receivers:\n  otlp:",
		Selector: map[string]string{
			"os.type": "linux",
		},
	}
	config.UpdateHash()
	if err := testStore.CreateConfiguration(ctx, config); err != nil {
		t.Fatalf("CreateConfiguration failed: %v", err)
	}

	// Get matching configuration
	retrieved, err := testStore.GetConfiguration(ctx, agent.ID)
	if err != nil {
		t.Fatalf("GetConfiguration failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected to find matching configuration")
	}
	if retrieved.Name != config.Name {
		t.Errorf("Configuration name = %v, want %v", retrieved.Name, config.Name)
	}
}

func TestStore_ListConfigurations(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Create configurations
	for i := 1; i <= 3; i++ {
		config := &model.Configuration{
			Name:        fmt.Sprintf("list-config-%03d", i),
			DisplayName: fmt.Sprintf("Config %d", i),
			ContentType: "yaml",
			RawConfig:   "receivers:\n  otlp:",
		}
		config.UpdateHash()
		if err := testStore.CreateConfiguration(ctx, config); err != nil {
			t.Fatalf("CreateConfiguration failed: %v", err)
		}
	}

	// List
	list, err := testStore.ListConfigurations(ctx)
	if err != nil {
		t.Fatalf("ListConfigurations failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("List length = %v, want 3", len(list))
	}
}

func TestStore_DeleteConfiguration(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	config := &model.Configuration{
		Name:        "delete-config",
		DisplayName: "Delete Test",
		ContentType: "yaml",
		RawConfig:   "receivers:\n  otlp:",
	}
	config.UpdateHash()

	if err := testStore.CreateConfiguration(ctx, config); err != nil {
		t.Fatalf("CreateConfiguration failed: %v", err)
	}

	// Delete
	if err := testStore.DeleteConfiguration(ctx, config.Name); err != nil {
		t.Fatalf("DeleteConfiguration failed: %v", err)
	}

	// Verify
	deletedConfig, err := testStore.GetConfigurationByName(ctx, config.Name)
	if err != nil || deletedConfig != nil {
		t.Errorf("Expected nil config and nil error for deleted configuration, got config=%v, err=%v", 
			deletedConfig, err)
	}
}
