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

	// 注意: Agent ID 在这时还不知道,需要等待第一条消息
	// 状态更新将在 onMessage 中处理
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

	// 更新心跳时间
	if err := s.store.UpdateAgentLastSeen(ctx, agentIDStr); err != nil {
		s.logger.Error("Failed to update agent last seen",
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

	ctx := context.Background()

	// 更新 Agent 状态为离线
	if err := s.store.UpdateAgentStatus(ctx, agentID, model.StatusOffline); err != nil {
		s.logger.Error("Failed to update agent status",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
		return
	}

	// 设置断开原因
	if err := s.store.SetAgentDisconnectReason(ctx, agentID, "connection closed"); err != nil {
		s.logger.Error("Failed to set disconnect reason",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
	}

	// 更新连接历史记录
	activeHistory, err := s.store.GetActiveConnectionHistory(ctx, agentID)
	if err != nil {
		s.logger.Error("Failed to get active connection history",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
		return
	}

	if activeHistory != nil {
		now := time.Now()
		activeHistory.DisconnectedAt = &now
		activeHistory.DisconnectReason = "connection closed"
		if err := s.store.UpdateConnectionHistory(ctx, activeHistory); err != nil {
			s.logger.Error("Failed to update connection history",
				zap.String("agent_id", agentID),
				zap.Error(err),
			)
		}
	}
}

// updateAgentState 更新 Agent 的状态信息
func (s *opampServer) updateAgentState(ctx context.Context, conn types.Connection, agentID string, message *protobufs.AgentToServer) error {
	// 获取或创建 Agent
	agent, err := s.store.GetAgent(ctx, agentID)
	if err != nil {
		return err
	}

	isNewAgent := (agent == nil)
	wasOffline := false

	if agent == nil {
		// Agent 不存在,创建新的
		agent = &model.Agent{
			ID:       agentID,
			Protocol: "opamp",
			Status:   model.StatusOffline, // 初始为离线,后面会更新
			Labels:   make(model.Labels),
		}
	} else {
		wasOffline = (agent.Status == model.StatusOffline)
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
	if wasOffline || isNewAgent {
		// Agent 从离线变为在线,更新状态
		agent.Status = model.StatusOnline
		agent.LastConnectedAt = &now
		agent.LastSeenAt = &now
		agent.DisconnectReason = ""

		// 注册连接到连接管理器
		s.connections.addConnection(agentID, conn)

		// 创建连接历史记录
		remoteAddr := "unknown"
		if conn.Connection() != nil {
			remoteAddr = conn.Connection().RemoteAddr().String()
		}

		history := &model.AgentConnectionHistory{
			AgentID:     agentID,
			ConnectedAt: now,
			RemoteAddr:  remoteAddr,
		}

		if err := s.store.CreateConnectionHistory(ctx, history); err != nil {
			s.logger.Error("Failed to create connection history",
				zap.String("agent_id", agentID),
				zap.Error(err),
			)
		}
	} else if agent.Status == model.StatusOnline {
		// 已经在线,只更新最后心跳时间
		agent.LastSeenAt = &now
	}

	// 检查配置状态
	if message.RemoteConfigStatus != nil {
		status := message.RemoteConfigStatus
		configHash := string(status.LastRemoteConfigHash)

		// Check if status indicates failure
		if status.Status == protobufs.RemoteConfigStatuses_RemoteConfigStatuses_FAILED {
			agent.Status = model.StatusError
			// 更新应用历史记录为失败状态
			s.updateApplyHistoryStatus(ctx, agentID, configHash, model.ApplyStatusFailed, status.ErrorMessage)
		} else if status.Status == protobufs.RemoteConfigStatuses_RemoteConfigStatuses_APPLIED {
			// 配置应用成功
			s.updateApplyHistoryStatus(ctx, agentID, configHash, model.ApplyStatusApplied, "")
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

// updateApplyHistoryStatus 更新配置应用历史状态
func (s *opampServer) updateApplyHistoryStatus(ctx context.Context, agentID, configHash string, status model.ApplyStatus, errorMsg string) {
	// 查找最近的待应用或应用中的记录
	histories, err := s.store.GetPendingApplyHistories(ctx)
	if err != nil {
		s.logger.Error("Failed to get pending apply histories",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
		return
	}

	// 查找匹配的记录
	for _, history := range histories {
		if history.AgentID == agentID && history.ConfigHash == configHash {
			history.Status = status
			if errorMsg != "" {
				history.ErrorMessage = errorMsg
			}
			if status == model.ApplyStatusApplied {
				now := time.Now()
				history.AppliedAt = &now
			}

			if err := s.store.UpdateApplyHistory(ctx, history); err != nil {
				s.logger.Error("Failed to update apply history",
					zap.String("agent_id", agentID),
					zap.Uint("history_id", history.ID),
					zap.Error(err),
				)
			}
			break
		}
	}
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
