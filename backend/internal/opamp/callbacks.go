package opamp

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"

	"github.com/cc1024201/opamp-platform/internal/model"
)

const (
	headerAuthorization = "Authorization"
	headerSecretKey     = "Secret-Key"
)

// onConnecting 在新连接建立前调用，用于验证和授权
func (s *opampServer) onConnecting(request *http.Request) types.ConnectionResponse {
	s.logger.Debug("Agent connecting",
		zap.String("remote_addr", request.RemoteAddr),
	)

	// 验证 Secret Key (如果配置了)
	if s.config.SecretKey != "" {
		secretKey := request.Header.Get(headerSecretKey)
		if secretKey == "" {
			// 尝试从 Authorization header 获取
			auth := request.Header.Get(headerAuthorization)
			if strings.HasPrefix(auth, "Bearer ") {
				secretKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if secretKey != s.config.SecretKey {
			s.logger.Warn("Invalid secret key",
				zap.String("remote_addr", request.RemoteAddr),
			)
			return types.ConnectionResponse{
				Accept:         false,
				HTTPStatusCode: http.StatusUnauthorized,
			}
		}
	}

	// 返回连接回调
	return types.ConnectionResponse{
		Accept:         true,
		HTTPStatusCode: http.StatusOK,
		ConnectionCallbacks: types.ConnectionCallbacks{
			OnConnected:       s.onConnected,
			OnMessage:         s.onMessage,
			OnConnectionClose: s.onConnectionClose,
		},
	}
}

// onConnected 在连接成功建立后调用
func (s *opampServer) onConnected(ctx context.Context, conn types.Connection) {
	remoteAddr := "unknown"
	if conn.Connection() != nil {
		remoteAddr = conn.Connection().RemoteAddr().String()
	}
	s.logger.Info("Agent connected", zap.String("remote_addr", remoteAddr))
}

// onMessage 处理从 Agent 接收到的消息
func (s *opampServer) onMessage(ctx context.Context, conn types.Connection, message *protobufs.AgentToServer) *protobufs.ServerToAgent {
	// 获取 Agent ID
	agentID := message.InstanceUid
	if len(agentID) == 0 {
		s.logger.Error("Agent message missing instance UID")
		return nil
	}

	agentIDStr := uuid.UUID(agentID).String()

	s.logger.Debug("Received message from agent",
		zap.String("agent_id", agentIDStr),
		zap.Uint64("sequence_num", message.SequenceNum),
	)

	// 更新 Agent 状态
	if err := s.updateAgentState(ctx, conn, agentIDStr, message); err != nil {
		s.logger.Error("Failed to update agent state",
			zap.String("agent_id", agentIDStr),
			zap.Error(err),
		)
	}

	// 检查是否需要发送配置
	response := s.checkAndSendConfig(ctx, agentIDStr, message)

	return response
}

// onConnectionClose 在连接关闭时调用
func (s *opampServer) onConnectionClose(conn types.Connection) {
	agentID := s.connections.removeConnection(conn)
	if agentID == "" {
		return
	}

	s.logger.Info("Agent disconnected", zap.String("agent_id", agentID))

	// 更新 Agent 状态为已断开
	ctx := context.Background()
	agent, err := s.store.GetAgent(ctx, agentID)
	if err != nil {
		s.logger.Error("Failed to get agent",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
		return
	}

	if agent == nil {
		return
	}

	now := time.Now()
	agent.Status = model.StatusDisconnected
	agent.DisconnectedAt = &now

	if err := s.store.UpsertAgent(ctx, agent); err != nil {
		s.logger.Error("Failed to update agent status",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
	}
}

// updateAgentState 更新 Agent 的状态信息
func (s *opampServer) updateAgentState(ctx context.Context, conn types.Connection, agentID string, message *protobufs.AgentToServer) error {
	// 获取或创建 Agent
	agent, err := s.store.GetAgent(ctx, agentID)
	if err != nil {
		return err
	}

	if agent == nil {
		// Agent 不存在，创建新的
		agent = &model.Agent{
			ID:       agentID,
			Protocol: "opamp",
			Status:   model.StatusConnected,
			Labels:   make(model.Labels),
		}
	}

	// 更新基本信息
	if message.AgentDescription != nil {
		desc := message.AgentDescription

		// 提取标识信息
		for _, attr := range desc.IdentifyingAttributes {
			switch attr.Key {
			case "service.name":
				agent.Name = attr.Value.GetStringValue()
			case "service.version":
				agent.Version = attr.Value.GetStringValue()
			case "host.name":
				agent.Hostname = attr.Value.GetStringValue()
			case "host.arch":
				agent.Architecture = attr.Value.GetStringValue()
			case "os.type":
				agent.Type = attr.Value.GetStringValue()
			}
		}

		// 提取非标识属性作为标签
		for _, attr := range desc.NonIdentifyingAttributes {
			agent.Labels[attr.Key] = attr.Value.GetStringValue()
		}
	}

	// 更新连接状态
	now := time.Now()
	if agent.Status == model.StatusDisconnected {
		agent.Status = model.StatusConnected
		agent.ConnectedAt = &now
		agent.DisconnectedAt = nil
	}

	// 检查配置状态
	if message.RemoteConfigStatus != nil {
		status := message.RemoteConfigStatus
		// Check if status indicates failure
		if status.Status == protobufs.RemoteConfigStatuses_RemoteConfigStatuses_FAILED {
			agent.Status = model.StatusError
		}
	}

	// 更新序列号
	agent.SequenceNumber = message.SequenceNum

	// 保存 Agent
	if err := s.store.UpsertAgent(ctx, agent); err != nil {
		return err
	}

	// 注册连接
	s.connections.addConnection(agentID, conn)

	return nil
}

// checkAndSendConfig 检查并发送配置给 Agent
func (s *opampServer) checkAndSendConfig(ctx context.Context, agentID string, message *protobufs.AgentToServer) *protobufs.ServerToAgent {
	// 获取 Agent 应该使用的配置
	config, err := s.store.GetConfiguration(ctx, agentID)
	if err != nil {
		s.logger.Error("Failed to get configuration for agent",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
		return nil
	}

	if config == nil {
		// 没有配置，不需要发送
		return nil
	}

	// 检查 Agent 当前配置的哈希
	var currentHash string
	if message.RemoteConfigStatus != nil {
		currentHash = string(message.RemoteConfigStatus.LastRemoteConfigHash)
	}

	// 如果配置相同，不需要发送
	if currentHash == config.ConfigHash {
		return nil
	}

	s.logger.Info("Sending new configuration to agent",
		zap.String("agent_id", agentID),
		zap.String("config_name", config.Name),
		zap.String("config_hash", config.ConfigHash),
	)

	// 构建配置消息
	response := &protobufs.ServerToAgent{
		InstanceUid: message.InstanceUid,
		RemoteConfig: &protobufs.AgentRemoteConfig{
			Config: &protobufs.AgentConfigMap{
				ConfigMap: map[string]*protobufs.AgentConfigFile{
					"config.yaml": {
						Body: []byte(config.RawConfig),
					},
				},
			},
			ConfigHash: []byte(config.ConfigHash),
		},
		Flags: uint64(protobufs.ServerToAgentFlags_ServerToAgentFlags_ReportFullState),
	}

	return response
}
