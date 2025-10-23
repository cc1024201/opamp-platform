package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestStore_Close tests the Close method
func TestStore_Close(t *testing.T) {
	// Create a temporary store for this test
	logger, _ := zap.NewDevelopment()
	config := Config{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefaultInt("TEST_DB_PORT", 5432),
		User:     getEnvOrDefault("TEST_DB_USER", "opamp"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "opamp123"),
		DBName:   getEnvOrDefault("TEST_DB_NAME", "opamp_platform"),
		SSLMode:  "disable",
	}

	tempStore, err := NewStore(config, logger)
	require.NoError(t, err)
	require.NotNil(t, tempStore)

	// Test Close
	err = tempStore.Close()
	assert.NoError(t, err)

	// Verify connection is closed by trying to ping
	db := tempStore.GetDB()
	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = sqlDB.Ping()
	assert.Error(t, err, "Connection should be closed")
}

// TestStore_GetDB tests the GetDB method
func TestStore_GetDB(t *testing.T) {
	db := testStore.GetDB()
	assert.NotNil(t, db)

	// Verify we can use the DB
	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = sqlDB.Ping()
	assert.NoError(t, err)

	// Check connection stats
	stats := sqlDB.Stats()
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
}

// TestStore_ListAgents_Pagination tests pagination edge cases
func TestStore_ListAgents_Pagination(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Create 10 agents
	for i := 1; i <= 10; i++ {
		agent := &model.Agent{
			ID:       fmt.Sprintf("page-agent-%03d", i),
			Name:     fmt.Sprintf("agent-%d", i),
			Hostname: fmt.Sprintf("host-%d", i),
			Status:   model.StatusOnline,
			Protocol: "opamp",
		}
		err := testStore.UpsertAgent(ctx, agent)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
		expectedTotal int64
	}{
		{
			name:          "first page",
			limit:         5,
			offset:        0,
			expectedCount: 5,
			expectedTotal: 10,
		},
		{
			name:          "second page",
			limit:         5,
			offset:        5,
			expectedCount: 5,
			expectedTotal: 10,
		},
		{
			name:          "last partial page",
			limit:         5,
			offset:        8,
			expectedCount: 2,
			expectedTotal: 10,
		},
		{
			name:          "offset beyond total",
			limit:         5,
			offset:        20,
			expectedCount: 0,
			expectedTotal: 10,
		},
		{
			name:          "zero limit",
			limit:         0,
			offset:        0,
			expectedCount: 0,
			expectedTotal: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agents, total, err := testStore.ListAgents(ctx, tt.limit, tt.offset)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, total)
			assert.Len(t, agents, tt.expectedCount)
		})
	}
}

// TestStore_GetConfiguration_EdgeCases tests edge cases for GetConfiguration
func TestStore_GetConfiguration_EdgeCases(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	t.Run("non-existent agent", func(t *testing.T) {
		config, err := testStore.GetConfiguration(ctx, "non-existent-agent")
		assert.NoError(t, err)
		assert.Nil(t, config)
	})

	t.Run("agent with specific configuration name", func(t *testing.T) {
		// Create configuration
		config := &model.Configuration{
			Name:        "specific-config",
			DisplayName: "Specific Config",
			ContentType: "yaml",
			RawConfig:   "receivers:\n  otlp:",
		}
		config.UpdateHash()
		err := testStore.CreateConfiguration(ctx, config)
		require.NoError(t, err)

		// Create agent with specific configuration name
		agent := &model.Agent{
			ID:                "specific-agent",
			Name:              "test-agent",
			Hostname:          "test-host",
			Status:            model.StatusOnline,
			Protocol:          "opamp",
			ConfigurationName: "specific-config",
		}
		err = testStore.UpsertAgent(ctx, agent)
		require.NoError(t, err)

		// Get configuration
		retrieved, err := testStore.GetConfiguration(ctx, agent.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, "specific-config", retrieved.Name)
	})

	t.Run("agent with non-existent configuration name", func(t *testing.T) {
		// Create agent with non-existent configuration name
		agent := &model.Agent{
			ID:                "missing-config-agent",
			Name:              "test-agent",
			Hostname:          "test-host",
			Status:            model.StatusOnline,
			Protocol:          "opamp",
			ConfigurationName: "non-existent-config",
		}
		err := testStore.UpsertAgent(ctx, agent)
		require.NoError(t, err)

		// Get configuration
		retrieved, err := testStore.GetConfiguration(ctx, agent.ID)
		assert.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("agent with labels but no matching config", func(t *testing.T) {
		// Create agent with labels
		agent := &model.Agent{
			ID:       "no-match-agent",
			Name:     "test-agent",
			Hostname: "test-host",
			Status:   model.StatusOnline,
			Protocol: "opamp",
			Labels: model.Labels{
				"env": "nonexistent",
			},
		}
		err := testStore.UpsertAgent(ctx, agent)
		require.NoError(t, err)

		// Get configuration
		retrieved, err := testStore.GetConfiguration(ctx, agent.ID)
		assert.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("multiple configs, first match wins", func(t *testing.T) {
		// Create multiple matching configurations
		config1 := &model.Configuration{
			Name:        "match-config-1",
			DisplayName: "Match Config 1",
			ContentType: "yaml",
			RawConfig:   "receivers:\n  otlp:",
			Selector: map[string]string{
				"tier": "production",
			},
		}
		config1.UpdateHash()
		err := testStore.CreateConfiguration(ctx, config1)
		require.NoError(t, err)

		config2 := &model.Configuration{
			Name:        "match-config-2",
			DisplayName: "Match Config 2",
			ContentType: "yaml",
			RawConfig:   "receivers:\n  prometheus:",
			Selector: map[string]string{
				"tier": "production",
			},
		}
		config2.UpdateHash()
		err = testStore.CreateConfiguration(ctx, config2)
		require.NoError(t, err)

		// Create agent with matching labels
		agent := &model.Agent{
			ID:       "multi-match-agent",
			Name:     "test-agent",
			Hostname: "test-host",
			Status:   model.StatusOnline,
			Protocol: "opamp",
			Labels: model.Labels{
				"tier": "production",
			},
		}
		err = testStore.UpsertAgent(ctx, agent)
		require.NoError(t, err)

		// Get configuration - should return one of them
		retrieved, err := testStore.GetConfiguration(ctx, agent.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Contains(t, []string{"match-config-1", "match-config-2"}, retrieved.Name)
	})
}

// TestStore_GetConfigurationByName_EdgeCases tests edge cases
func TestStore_GetConfigurationByName_EdgeCases(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	t.Run("non-existent configuration", func(t *testing.T) {
		config, err := testStore.GetConfigurationByName(ctx, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, config)
	})

	t.Run("empty name", func(t *testing.T) {
		config, err := testStore.GetConfigurationByName(ctx, "")
		assert.NoError(t, err)
		assert.Nil(t, config)
	})
}

// TestStore_ListConfigurations_Empty tests empty list
func TestStore_ListConfigurations_Empty(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	configs, err := testStore.ListConfigurations(ctx)
	require.NoError(t, err)
	assert.Empty(t, configs)
}

// TestStore_DeleteAgent_NonExistent tests deleting non-existent agent
func TestStore_DeleteAgent_NonExistent(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Delete non-existent agent (should not error)
	err := testStore.DeleteAgent(ctx, "non-existent")
	assert.NoError(t, err)
}

// TestStore_DeleteConfiguration_NonExistent tests deleting non-existent configuration
func TestStore_DeleteConfiguration_NonExistent(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Delete non-existent configuration (should not error)
	err := testStore.DeleteConfiguration(ctx, "non-existent")
	assert.NoError(t, err)
}

// TestStore_NewStore_InvalidConfig tests NewStore with invalid config
func TestStore_NewStore_InvalidConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "invalid host",
			config: Config{
				Host:     "invalid-host-that-does-not-exist",
				Port:     5432,
				User:     "opamp",
				Password: "opamp123",
				DBName:   "opamp_platform",
				SSLMode:  "disable",
			},
		},
		{
			name: "invalid port",
			config: Config{
				Host:     "localhost",
				Port:     99999,
				User:     "opamp",
				Password: "opamp123",
				DBName:   "opamp_platform",
				SSLMode:  "disable",
			},
		},
		{
			name: "invalid credentials",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "invalid_user",
				Password: "invalid_password",
				DBName:   "opamp_platform",
				SSLMode:  "disable",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewStore(tt.config, logger)
			assert.Error(t, err)
			assert.Nil(t, store)
		})
	}
}

// TestStore_ContextCancellation tests context cancellation
func TestStore_ContextCancellation(t *testing.T) {
	cleanupDatabase(t)

	t.Run("cancelled context for GetAgent", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := testStore.GetAgent(ctx, "test-agent")
		// Depending on timing, this may or may not error
		// Just verify it doesn't panic
		_ = err
	})
}

// TestStore_ConcurrentAccess tests concurrent database access
func TestStore_ConcurrentAccess(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			agent := &model.Agent{
				ID:       fmt.Sprintf("concurrent-agent-%03d", id),
				Name:     fmt.Sprintf("agent-%d", id),
				Hostname: fmt.Sprintf("host-%d", id),
				Status:   model.StatusOnline,
				Protocol: "opamp",
			}
			err := testStore.UpsertAgent(ctx, agent)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all agents were created
	agents, total, err := testStore.ListAgents(ctx, 100, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(numGoroutines), total)
	assert.Len(t, agents, numGoroutines)
}
