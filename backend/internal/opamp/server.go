package opamp

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opamp-go/server"
	"github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"

	"github.com/cc1024201/opamp-platform/internal/model"
)

// Server 定义 OpAMP 服务器接口
type Server interface {
	// Start 启动 OpAMP 服务器
	Start(ctx context.Context) error
	// Stop 停止 OpAMP 服务器
	Stop(ctx context.Context) error
	// Handler 返回 HTTP 处理函数
	Handler() http.HandlerFunc
	// Connected 检查 Agent 是否已连接
	Connected(agentID string) bool
	// SendUpdate 向 Agent 发送更新
	SendUpdate(ctx context.Context, agentID string, update *model.AgentUpdate) error
}

// Config OpAMP 服务器配置
type Config struct {
	Endpoint  string // OpAMP 端点路径
	SecretKey string // Secret Key (为空则不验证)
}

// AgentStore 定义 Agent 存储接口
type AgentStore interface {
	// GetAgent 获取 Agent
	GetAgent(ctx context.Context, agentID string) (*model.Agent, error)
	// UpsertAgent 创建或更新 Agent
	UpsertAgent(ctx context.Context, agent *model.Agent) error
	// GetConfiguration 获取 Agent 的配置
	GetConfiguration(ctx context.Context, agentID string) (*model.Configuration, error)
	// GetPendingApplyHistories 获取所有待应用或应用中的配置记录
	GetPendingApplyHistories(ctx context.Context) ([]*model.ConfigurationApplyHistory, error)
	// UpdateApplyHistory 更新配置应用历史
	UpdateApplyHistory(ctx context.Context, history *model.ConfigurationApplyHistory) error
}

type opampServer struct {
	config      Config
	logger      *zap.Logger
	server      server.OpAMPServer
	handler     server.HTTPHandlerFunc
	store       AgentStore
	connections *connectionManager
}

// NewServer 创建新的 OpAMP 服务器
func NewServer(config Config, store AgentStore, logger *zap.Logger) (Server, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	s := &opampServer{
		config:      config,
		logger:      logger,
		store:       store,
		connections: newConnectionManager(),
	}

	// 创建 opamp-go 服务器
	opampServer := server.New(newLoggerAdapter(logger))

	settings := server.Settings{
		Callbacks: types.Callbacks{
			OnConnecting: s.onConnecting,
		},
	}

	handler, _, err := opampServer.Attach(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to attach OpAMP server: %w", err)
	}

	s.server = opampServer
	s.handler = handler
	return s, nil
}

func (s *opampServer) Start(ctx context.Context) error {
	s.logger.Info("OpAMP server started", zap.String("endpoint", s.config.Endpoint))
	return nil
}

func (s *opampServer) Stop(ctx context.Context) error {
	s.logger.Info("OpAMP server stopping")
	return s.server.Stop(ctx)
}

func (s *opampServer) Handler() http.HandlerFunc {
	return http.HandlerFunc(s.handler)
}

func (s *opampServer) Connected(agentID string) bool {
	return s.connections.isConnected(agentID)
}

func (s *opampServer) SendUpdate(ctx context.Context, agentID string, update *model.AgentUpdate) error {
	conn := s.connections.getConnection(agentID)
	if conn == nil {
		return fmt.Errorf("agent %s not connected", agentID)
	}

	// 构建 ServerToAgent 消息
	msg := &protobufs.ServerToAgent{}

	// 如果有配置更新
	if update.Configuration != nil {
		msg.RemoteConfig = &protobufs.AgentRemoteConfig{
			Config: &protobufs.AgentConfigMap{
				ConfigMap: map[string]*protobufs.AgentConfigFile{
					"config.yaml": {
						Body: []byte(update.Configuration.RawConfig),
					},
				},
			},
		}
		// 计算配置哈希
		msg.RemoteConfig.ConfigHash = []byte(update.Configuration.ConfigHash)
	}

	// 发送消息
	return conn.Send(ctx, msg)
}

// connectionManager 管理 Agent 连接
type connectionManager struct {
	mu          sync.RWMutex
	connections map[string]types.Connection // agentID -> connection
	agents      map[types.Connection]string // connection -> agentID
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
		connections: make(map[string]types.Connection),
		agents:      make(map[types.Connection]string),
	}
}

func (cm *connectionManager) addConnection(agentID string, conn types.Connection) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.connections[agentID] = conn
	cm.agents[conn] = agentID
}

func (cm *connectionManager) removeConnection(conn types.Connection) string {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	agentID := cm.agents[conn]
	if agentID != "" {
		delete(cm.connections, agentID)
		delete(cm.agents, conn)
	}
	return agentID
}

func (cm *connectionManager) getConnection(agentID string) types.Connection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.connections[agentID]
}

func (cm *connectionManager) isConnected(agentID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.connections[agentID] != nil
}
