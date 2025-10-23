package opamp

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/open-telemetry/opamp-go/protobufs"
	"go.uber.org/zap"

	"github.com/cc1024201/opamp-platform/internal/model"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	remoteAddr net.Addr
}

func (m *mockConn) Read(b []byte) (n int, err error)      { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)     { return len(b), nil }
func (m *mockConn) Close() error                          { return nil }
func (m *mockConn) LocalAddr() net.Addr                   { return m.remoteAddr }
func (m *mockConn) RemoteAddr() net.Addr                  { return m.remoteAddr }
func (m *mockConn) SetDeadline(t time.Time) error         { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error     { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error    { return nil }

// mockAddr implements net.Addr for testing
type mockAddr struct {
	addr string
}

func (m *mockAddr) Network() string { return "tcp" }
func (m *mockAddr) String() string  { return m.addr }

// mockConnection implements types.Connection for testing
type mockConnection struct {
	id   string
	conn net.Conn
}

func newMockConnection(id string) *mockConnection {
	return &mockConnection{
		id: id,
		conn: &mockConn{
			remoteAddr: &mockAddr{addr: "127.0.0.1:12345"},
		},
	}
}

func (m *mockConnection) Send(ctx context.Context, message *protobufs.ServerToAgent) error {
	return nil
}

func (m *mockConnection) Connection() net.Conn {
	return m.conn
}

func (m *mockConnection) Disconnect() error {
	return nil
}

// mockAgentStore implements AgentStore for testing
type mockAgentStore struct {
	mu            sync.RWMutex
	agents        map[string]*model.Agent
	configurations map[string]*model.Configuration
	getAgentErr   error
	upsertErr     error
	getConfigErr  error
}

func newMockAgentStore() *mockAgentStore {
	return &mockAgentStore{
		agents:         make(map[string]*model.Agent),
		configurations: make(map[string]*model.Configuration),
	}
}

func (m *mockAgentStore) GetAgent(ctx context.Context, agentID string) (*model.Agent, error) {
	if m.getAgentErr != nil {
		return nil, m.getAgentErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.agents[agentID], nil
}

func (m *mockAgentStore) UpsertAgent(ctx context.Context, agent *model.Agent) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[agent.ID] = agent
	return nil
}

func (m *mockAgentStore) GetConfiguration(ctx context.Context, agentID string) (*model.Configuration, error) {
	if m.getConfigErr != nil {
		return nil, m.getConfigErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.configurations[agentID], nil
}

func (m *mockAgentStore) GetPendingApplyHistories(ctx context.Context) ([]*model.ConfigurationApplyHistory, error) {
	// Mock implementation - returns empty slice
	return []*model.ConfigurationApplyHistory{}, nil
}

func (m *mockAgentStore) UpdateApplyHistory(ctx context.Context, history *model.ConfigurationApplyHistory) error {
	// Mock implementation - does nothing
	return nil
}

// Agent 状态管理方法
func (m *mockAgentStore) UpdateAgentStatus(ctx context.Context, agentID string, status model.AgentStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if agent, exists := m.agents[agentID]; exists {
		agent.Status = status
		now := time.Now()
		if status == model.StatusOnline {
			agent.LastConnectedAt = &now
		} else if status == model.StatusOffline {
			agent.LastDisconnectedAt = &now
		}
	}
	return nil
}

func (m *mockAgentStore) UpdateAgentLastSeen(ctx context.Context, agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if agent, exists := m.agents[agentID]; exists {
		now := time.Now()
		agent.LastSeenAt = &now
	}
	return nil
}

func (m *mockAgentStore) SetAgentDisconnectReason(ctx context.Context, agentID string, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if agent, exists := m.agents[agentID]; exists {
		agent.DisconnectReason = reason
	}
	return nil
}

func (m *mockAgentStore) ListStaleAgents(ctx context.Context, timeout time.Duration) ([]*model.Agent, error) {
	return []*model.Agent{}, nil
}

// 连接历史管理方法
func (m *mockAgentStore) CreateConnectionHistory(ctx context.Context, history *model.AgentConnectionHistory) error {
	return nil
}

func (m *mockAgentStore) UpdateConnectionHistory(ctx context.Context, history *model.AgentConnectionHistory) error {
	return nil
}

func (m *mockAgentStore) GetActiveConnectionHistory(ctx context.Context, agentID string) (*model.AgentConnectionHistory, error) {
	return nil, nil
}

func TestNewServer(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()

	tests := []struct {
		name    string
		config  Config
		store   AgentStore
		logger  *zap.Logger
		wantErr bool
	}{
		{
			name: "valid config with logger",
			config: Config{
				Endpoint:  "/v1/opamp",
				SecretKey: "test-secret",
			},
			store:   store,
			logger:  logger,
			wantErr: false,
		},
		{
			name: "valid config with nil logger",
			config: Config{
				Endpoint: "/v1/opamp",
			},
			store:   store,
			logger:  nil,
			wantErr: false,
		},
		{
			name: "valid config without secret key",
			config: Config{
				Endpoint: "/v1/opamp",
			},
			store:   store,
			logger:  logger,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewServer(tt.config, tt.store, tt.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && server == nil {
				t.Error("NewServer() returned nil server")
			}
		})
	}
}

func TestConnectionManager_AddConnection(t *testing.T) {
	cm := newConnectionManager()
	conn := newMockConnection("conn-1")
	agentID := "agent-001"

	cm.addConnection(agentID, conn)

	// Verify connection was added
	if !cm.isConnected(agentID) {
		t.Error("Expected agent to be connected")
	}

	retrievedConn := cm.getConnection(agentID)
	if retrievedConn == nil {
		t.Error("Retrieved connection should not be nil")
	}
}

func TestConnectionManager_RemoveConnection(t *testing.T) {
	cm := newConnectionManager()
	conn := newMockConnection("conn-1")
	agentID := "agent-001"

	cm.addConnection(agentID, conn)

	// Remove connection
	removedID := cm.removeConnection(conn)

	if removedID != agentID {
		t.Errorf("removeConnection() = %v, want %v", removedID, agentID)
	}

	// Verify connection was removed
	if cm.isConnected(agentID) {
		t.Error("Expected agent to be disconnected")
	}

	if cm.getConnection(agentID) != nil {
		t.Error("Expected nil connection after removal")
	}
}

func TestConnectionManager_Concurrent(t *testing.T) {
	cm := newConnectionManager()
	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			agentID := string(rune(id))
			conn := newMockConnection(agentID)
			cm.addConnection(agentID, conn)
		}(i)
	}

	wg.Wait()

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			agentID := string(rune(id))
			cm.isConnected(agentID)
			cm.getConnection(agentID)
		}(i)
	}

	wg.Wait()
}

func TestOnConnecting_NoSecretKey(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{
		Endpoint:  "/v1/opamp",
		SecretKey: "", // No secret key
	}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)

	req := httptest.NewRequest("GET", "/v1/opamp", nil)
	resp := opampSrv.onConnecting(req)

	if !resp.Accept {
		t.Error("Expected connection to be accepted when no secret key is configured")
	}
	if resp.HTTPStatusCode != http.StatusOK {
		t.Errorf("HTTPStatusCode = %v, want %v", resp.HTTPStatusCode, http.StatusOK)
	}
}

func TestOnConnecting_ValidSecretKey(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	secretKey := "test-secret-key"
	config := Config{
		Endpoint:  "/v1/opamp",
		SecretKey: secretKey,
	}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)

	tests := []struct {
		name           string
		headerName     string
		headerValue    string
		wantAccept     bool
		wantStatusCode int
	}{
		{
			name:           "valid secret key in Secret-Key header",
			headerName:     headerSecretKey,
			headerValue:    secretKey,
			wantAccept:     true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "valid secret key in Authorization Bearer",
			headerName:     headerAuthorization,
			headerValue:    "Bearer " + secretKey,
			wantAccept:     true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid secret key",
			headerName:     headerSecretKey,
			headerValue:    "wrong-secret",
			wantAccept:     false,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "missing secret key",
			headerName:     "",
			headerValue:    "",
			wantAccept:     false,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/opamp", nil)
			if tt.headerName != "" {
				req.Header.Set(tt.headerName, tt.headerValue)
			}

			resp := opampSrv.onConnecting(req)

			if resp.Accept != tt.wantAccept {
				t.Errorf("Accept = %v, want %v", resp.Accept, tt.wantAccept)
			}
			if resp.HTTPStatusCode != tt.wantStatusCode {
				t.Errorf("HTTPStatusCode = %v, want %v", resp.HTTPStatusCode, tt.wantStatusCode)
			}

			// Verify callbacks are set when accepted
			if resp.Accept {
				if resp.ConnectionCallbacks.OnConnected == nil {
					t.Error("OnConnected callback should not be nil")
				}
				if resp.ConnectionCallbacks.OnMessage == nil {
					t.Error("OnMessage callback should not be nil")
				}
				if resp.ConnectionCallbacks.OnConnectionClose == nil {
					t.Error("OnConnectionClose callback should not be nil")
				}
			}
		})
	}
}

func TestConnected(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)
	agentID := "test-agent-001"
	conn := newMockConnection("conn-1")

	// Initially not connected
	if opampSrv.Connected(agentID) {
		t.Error("Expected agent to not be connected initially")
	}

	// Add connection
	opampSrv.connections.addConnection(agentID, conn)

	// Now should be connected
	if !opampSrv.Connected(agentID) {
		t.Error("Expected agent to be connected after adding connection")
	}

	// Remove connection
	opampSrv.connections.removeConnection(conn)

	// Should not be connected anymore
	if opampSrv.Connected(agentID) {
		t.Error("Expected agent to not be connected after removal")
	}
}

func TestSendUpdate_AgentNotConnected(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	ctx := context.Background()
	update := &model.AgentUpdate{}

	err = server.SendUpdate(ctx, "non-existent-agent", update)
	if err == nil {
		t.Error("Expected error when sending update to non-connected agent")
	}
}

func TestSendUpdate_WithConfiguration(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	opampSrv := server.(*opampServer)
	ctx := context.Background()
	agentID := "test-agent-001"
	conn := newMockConnection("conn-1")

	// Add connection
	opampSrv.connections.addConnection(agentID, conn)

	// Create update with configuration
	configuration := &model.Configuration{
		Name:       "test-config",
		RawConfig:  "receivers:\n  otlp:",
		ConfigHash: "abc123",
	}
	update := &model.AgentUpdate{
		Configuration: configuration,
	}

	err = server.SendUpdate(ctx, agentID, update)
	if err != nil {
		t.Errorf("SendUpdate() failed: %v", err)
	}
}

func TestHandler(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	handler := server.Handler()
	if handler == nil {
		t.Error("Handler() returned nil")
	}
}

func TestStartStop(t *testing.T) {
	logger := zap.NewNop()
	store := newMockAgentStore()
	config := Config{Endpoint: "/v1/opamp"}

	server, err := NewServer(config, store, logger)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	ctx := context.Background()

	// Test Start
	err = server.Start(ctx)
	if err != nil {
		t.Errorf("Start() failed: %v", err)
	}

	// Test Stop
	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Stop() failed: %v", err)
	}
}
