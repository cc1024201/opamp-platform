package model

import (
	"testing"
	"time"
)

func TestLabels_Matches(t *testing.T) {
	tests := []struct {
		name     string
		labels   Labels
		selector map[string]string
		want     bool
	}{
		{
			name: "empty selector matches any labels",
			labels: Labels{
				"env":    "prod",
				"region": "us-east",
			},
			selector: map[string]string{},
			want:     true,
		},
		{
			name: "exact match single label",
			labels: Labels{
				"env":    "prod",
				"region": "us-east",
			},
			selector: map[string]string{
				"env": "prod",
			},
			want: true,
		},
		{
			name: "exact match multiple labels",
			labels: Labels{
				"env":    "prod",
				"region": "us-east",
				"tier":   "frontend",
			},
			selector: map[string]string{
				"env":    "prod",
				"region": "us-east",
			},
			want: true,
		},
		{
			name: "no match - wrong value",
			labels: Labels{
				"env":    "prod",
				"region": "us-east",
			},
			selector: map[string]string{
				"env": "dev",
			},
			want: false,
		},
		{
			name: "no match - missing label",
			labels: Labels{
				"env": "prod",
			},
			selector: map[string]string{
				"region": "us-east",
			},
			want: false,
		},
		{
			name: "subset match - selector is subset of labels",
			labels: Labels{
				"env":      "prod",
				"region":   "us-east",
				"tier":     "frontend",
				"version":  "v1.0.0",
				"hostname": "web-01",
			},
			selector: map[string]string{
				"env":  "prod",
				"tier": "frontend",
			},
			want: true,
		},
		{
			name:     "empty labels with non-empty selector",
			labels:   Labels{},
			selector: map[string]string{"env": "prod"},
			want:     false,
		},
		{
			name:     "nil labels with non-empty selector",
			labels:   nil,
			selector: map[string]string{"env": "prod"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.labels.Matches(tt.selector)
			if got != tt.want {
				t.Errorf("Labels.Matches() = %v, want %v", got, tt.want)
				t.Logf("Labels: %+v", tt.labels)
				t.Logf("Selector: %+v", tt.selector)
			}
		})
	}
}

func TestAgent_Status(t *testing.T) {
	tests := []struct {
		name   string
		status AgentStatus
		want   string
	}{
		{"Disconnected", StatusDisconnected, "Disconnected"},
		{"Connected", StatusConnected, "Connected"},
		{"Configuring", StatusConfiguring, "Configuring"},
		{"Error", StatusError, "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ID:     "test-agent-id",
				Status: tt.status,
			}
			if agent.Status != tt.status {
				t.Errorf("Agent.Status = %v, want %v", agent.Status, tt.status)
			}
		})
	}
}

func TestAgent_Creation(t *testing.T) {
	now := time.Now()
	agent := &Agent{
		ID:           "test-id-123",
		Name:         "test-collector",
		Type:         "otel-collector",
		Architecture: "amd64",
		Hostname:     "test-host",
		Version:      "1.0.0",
		Status:       StatusConnected,
		ConnectedAt:  &now,
		Labels: Labels{
			"env":    "test",
			"region": "us-west",
		},
		Protocol:       "opamp",
		SequenceNumber: 1,
	}

	// Validate basic fields
	if agent.ID != "test-id-123" {
		t.Errorf("Agent.ID = %v, want %v", agent.ID, "test-id-123")
	}
	if agent.Name != "test-collector" {
		t.Errorf("Agent.Name = %v, want %v", agent.Name, "test-collector")
	}
	if agent.Status != StatusConnected {
		t.Errorf("Agent.Status = %v, want %v", agent.Status, StatusConnected)
	}
	if agent.ConnectedAt == nil {
		t.Error("Agent.ConnectedAt should not be nil")
	}
	if len(agent.Labels) != 2 {
		t.Errorf("Agent.Labels length = %v, want %v", len(agent.Labels), 2)
	}
}

func TestAgent_Labels(t *testing.T) {
	agent := &Agent{
		ID: "test-agent",
		Labels: Labels{
			"os.type":   "linux",
			"host.name": "server-01",
			"env":       "production",
		},
	}

	// Test label access
	if agent.Labels["os.type"] != "linux" {
		t.Errorf("Expected os.type=linux, got %v", agent.Labels["os.type"])
	}

	// Test label matching
	selector1 := map[string]string{"os.type": "linux"}
	if !agent.Labels.Matches(selector1) {
		t.Error("Agent labels should match selector with os.type=linux")
	}

	selector2 := map[string]string{"os.type": "windows"}
	if agent.Labels.Matches(selector2) {
		t.Error("Agent labels should not match selector with os.type=windows")
	}

	selector3 := map[string]string{
		"os.type": "linux",
		"env":     "production",
	}
	if !agent.Labels.Matches(selector3) {
		t.Error("Agent labels should match multi-label selector")
	}
}

func TestAgent_ConfigurationName(t *testing.T) {
	tests := []struct {
		name              string
		configurationName string
		wantEmpty         bool
	}{
		{
			name:              "agent with explicit configuration",
			configurationName: "prod-config",
			wantEmpty:         false,
		},
		{
			name:              "agent without configuration",
			configurationName: "",
			wantEmpty:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				ID:                "test-agent",
				ConfigurationName: tt.configurationName,
			}

			isEmpty := agent.ConfigurationName == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("ConfigurationName empty = %v, want %v", isEmpty, tt.wantEmpty)
			}
		})
	}
}
