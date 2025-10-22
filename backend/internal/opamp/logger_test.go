package opamp

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerAdapter_Debugf(t *testing.T) {
	// Create observed logger to capture log output
	core, observed := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	adapter := newLoggerAdapter(logger)

	ctx := context.Background()
	adapter.Debugf(ctx, "test debug message: %s", "value")

	// Verify log was written
	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	entry := logs[0]
	if entry.Level != zap.DebugLevel {
		t.Errorf("Log level = %v, want %v", entry.Level, zap.DebugLevel)
	}

	if entry.Message != "test debug message: value" {
		t.Errorf("Log message = %v, want 'test debug message: value'", entry.Message)
	}
}

func TestLoggerAdapter_Errorf(t *testing.T) {
	// Create observed logger to capture log output
	core, observed := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	adapter := newLoggerAdapter(logger)

	ctx := context.Background()
	adapter.Errorf(ctx, "test error message: %d", 42)

	// Verify log was written
	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	entry := logs[0]
	if entry.Level != zap.ErrorLevel {
		t.Errorf("Log level = %v, want %v", entry.Level, zap.ErrorLevel)
	}

	if entry.Message != "test error message: 42" {
		t.Errorf("Log message = %v, want 'test error message: 42'", entry.Message)
	}
}

func TestNewLoggerAdapter(t *testing.T) {
	logger := zap.NewNop()
	adapter := newLoggerAdapter(logger)

	if adapter == nil {
		t.Error("newLoggerAdapter() returned nil")
	}

	// Verify it implements the Logger interface by calling methods
	ctx := context.Background()
	adapter.Debugf(ctx, "test")
	adapter.Errorf(ctx, "test")
}
