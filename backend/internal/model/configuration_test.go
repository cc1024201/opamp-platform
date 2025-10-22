package model

import (
	"testing"
)

func TestConfiguration_UpdateHash(t *testing.T) {
	config := &Configuration{
		Name:      "test-config",
		RawConfig: "receivers:\n  otlp:\nexporters:\n  logging:",
	}

	// Update hash
	config.UpdateHash()

	// Hash should not be empty
	if config.ConfigHash == "" {
		t.Error("ConfigHash should not be empty after UpdateHash()")
	}

	// Hash should be 64 characters (SHA-256 hex)
	if len(config.ConfigHash) != 64 {
		t.Errorf("ConfigHash length = %d, want 64", len(config.ConfigHash))
	}

	// Save the hash
	firstHash := config.ConfigHash

	// Update hash again with same content - should produce same hash
	config.UpdateHash()
	if config.ConfigHash != firstHash {
		t.Errorf("ConfigHash changed on re-calculation: %s != %s", config.ConfigHash, firstHash)
	}

	// Change content - should produce different hash
	config.RawConfig = "receivers:\n  otlp:\nexporters:\n  otlp:"
	config.UpdateHash()
	if config.ConfigHash == firstHash {
		t.Error("ConfigHash should change when RawConfig changes")
	}
}

func TestConfiguration_MatchesAgent(t *testing.T) {
	tests := []struct {
		name     string
		config   *Configuration
		agent    *Agent
		want     bool
		reason   string
	}{
		{
			name: "exact match single label",
			config: &Configuration{
				Name: "linux-config",
				Selector: map[string]string{
					"os.type": "linux",
				},
			},
			agent: &Agent{
				ID: "agent-1",
				Labels: Labels{
					"os.type":   "linux",
					"host.name": "server-01",
				},
			},
			want:   true,
			reason: "agent has os.type=linux",
		},
		{
			name: "exact match multiple labels",
			config: &Configuration{
				Name: "prod-linux-config",
				Selector: map[string]string{
					"os.type": "linux",
					"env":     "production",
				},
			},
			agent: &Agent{
				ID: "agent-2",
				Labels: Labels{
					"os.type":   "linux",
					"env":       "production",
					"host.name": "prod-server-01",
				},
			},
			want:   true,
			reason: "agent has both os.type=linux and env=production",
		},
		{
			name: "no match - wrong label value",
			config: &Configuration{
				Name: "windows-config",
				Selector: map[string]string{
					"os.type": "windows",
				},
			},
			agent: &Agent{
				ID: "agent-3",
				Labels: Labels{
					"os.type": "linux",
				},
			},
			want:   false,
			reason: "agent has os.type=linux, not windows",
		},
		{
			name: "no match - missing label",
			config: &Configuration{
				Name: "eu-config",
				Selector: map[string]string{
					"region": "eu-west",
				},
			},
			agent: &Agent{
				ID: "agent-4",
				Labels: Labels{
					"os.type": "linux",
					"env":     "production",
				},
			},
			want:   false,
			reason: "agent does not have region label",
		},
		{
			name: "empty selector - no match",
			config: &Configuration{
				Name:     "no-selector-config",
				Selector: map[string]string{},
			},
			agent: &Agent{
				ID: "agent-5",
				Labels: Labels{
					"os.type": "linux",
				},
			},
			want:   false,
			reason: "empty selector should not match",
		},
		{
			name: "nil selector - no match",
			config: &Configuration{
				Name:     "nil-selector-config",
				Selector: nil,
			},
			agent: &Agent{
				ID: "agent-6",
				Labels: Labels{
					"os.type": "linux",
				},
			},
			want:   false,
			reason: "nil selector should not match",
		},
		{
			name: "partial match - selector is subset",
			config: &Configuration{
				Name: "linux-config",
				Selector: map[string]string{
					"os.type": "linux",
				},
			},
			agent: &Agent{
				ID: "agent-7",
				Labels: Labels{
					"os.type":   "linux",
					"env":       "production",
					"region":    "us-east",
					"host.name": "web-01",
				},
			},
			want:   true,
			reason: "agent has all required labels (os.type=linux) plus extra labels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.MatchesAgent(tt.agent)
			if got != tt.want {
				t.Errorf("Configuration.MatchesAgent() = %v, want %v", got, tt.want)
				t.Logf("Reason: %s", tt.reason)
				t.Logf("Config Selector: %+v", tt.config.Selector)
				t.Logf("Agent Labels: %+v", tt.agent.Labels)
			}
		})
	}
}

func TestConfiguration_HashStability(t *testing.T) {
	// Test that hash is stable across multiple calculations
	config := &Configuration{
		Name:        "stable-config",
		DisplayName: "Stable Configuration",
		RawConfig: `receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
exporters:
  logging:
    loglevel: debug`,
	}

	// Calculate hash multiple times
	hashes := make([]string, 5)
	for i := 0; i < 5; i++ {
		config.UpdateHash()
		hashes[i] = config.ConfigHash
	}

	// All hashes should be identical
	firstHash := hashes[0]
	for i, hash := range hashes {
		if hash != firstHash {
			t.Errorf("Hash %d differs: %s != %s", i, hash, firstHash)
		}
	}
}

func TestConfiguration_Creation(t *testing.T) {
	config := &Configuration{
		Name:        "test-config",
		DisplayName: "Test Configuration",
		Description: "A test configuration",
		ContentType: "yaml",
		RawConfig:   "receivers:\n  otlp:",
		Selector: map[string]string{
			"env": "test",
		},
	}

	// Validate fields
	if config.Name != "test-config" {
		t.Errorf("Name = %v, want %v", config.Name, "test-config")
	}
	if config.DisplayName != "Test Configuration" {
		t.Errorf("DisplayName = %v, want %v", config.DisplayName, "Test Configuration")
	}
	if config.ContentType != "yaml" {
		t.Errorf("ContentType = %v, want %v", config.ContentType, "yaml")
	}
	if len(config.Selector) != 1 {
		t.Errorf("Selector length = %v, want 1", len(config.Selector))
	}

	// Generate hash
	config.UpdateHash()
	if config.ConfigHash == "" {
		t.Error("ConfigHash should not be empty")
	}
}

func TestSource_Creation(t *testing.T) {
	source := &Source{
		Name: "test-source",
		Type: "otlp",
		Parameters: map[string]interface{}{
			"endpoint": "0.0.0.0:4317",
			"protocol": "grpc",
		},
	}

	if source.Name != "test-source" {
		t.Errorf("Name = %v, want %v", source.Name, "test-source")
	}
	if source.Type != "otlp" {
		t.Errorf("Type = %v, want %v", source.Type, "otlp")
	}
	if len(source.Parameters) != 2 {
		t.Errorf("Parameters length = %v, want 2", len(source.Parameters))
	}
}

func TestDestination_Creation(t *testing.T) {
	dest := &Destination{
		Name: "test-destination",
		Type: "logging",
		Parameters: map[string]interface{}{
			"loglevel": "debug",
		},
	}

	if dest.Name != "test-destination" {
		t.Errorf("Name = %v, want %v", dest.Name, "test-destination")
	}
	if dest.Type != "logging" {
		t.Errorf("Type = %v, want %v", dest.Type, "logging")
	}
}

func TestProcessor_Creation(t *testing.T) {
	processor := &Processor{
		Name: "test-processor",
		Type: "batch",
		Parameters: map[string]interface{}{
			"timeout":    "10s",
			"batch_size": 100,
		},
	}

	if processor.Name != "test-processor" {
		t.Errorf("Name = %v, want %v", processor.Name, "test-processor")
	}
	if processor.Type != "batch" {
		t.Errorf("Type = %v, want %v", processor.Type, "batch")
	}
	if len(processor.Parameters) != 2 {
		t.Errorf("Parameters length = %v, want 2", len(processor.Parameters))
	}
}

func TestConfiguration_SelectorValidation(t *testing.T) {
	tests := []struct {
		name     string
		selector map[string]string
		wantNil  bool
	}{
		{
			name:     "valid selector",
			selector: map[string]string{"env": "prod"},
			wantNil:  false,
		},
		{
			name:     "empty selector",
			selector: map[string]string{},
			wantNil:  false,
		},
		{
			name:     "nil selector",
			selector: nil,
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Configuration{
				Name:     "test",
				Selector: tt.selector,
			}

			isNil := config.Selector == nil
			if isNil != tt.wantNil {
				t.Errorf("Selector nil = %v, want %v", isNil, tt.wantNil)
			}
		})
	}
}
