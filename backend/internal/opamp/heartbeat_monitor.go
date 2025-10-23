package opamp

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/cc1024201/opamp-platform/internal/metrics"
	"github.com/cc1024201/opamp-platform/internal/model"
)

// HeartbeatMonitor 心跳监控器
type HeartbeatMonitor struct {
	store         AgentStore
	logger        *zap.Logger
	metrics       *metrics.Metrics
	checkInterval time.Duration
	timeout       time.Duration
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// NewHeartbeatMonitor 创建新的心跳监控器
func NewHeartbeatMonitor(store AgentStore, logger *zap.Logger, m *metrics.Metrics, checkInterval, timeout time.Duration) *HeartbeatMonitor {
	if logger == nil {
		logger = zap.NewNop()
	}

	// 默认值
	if checkInterval == 0 {
		checkInterval = 30 * time.Second // 每 30 秒检查一次
	}
	if timeout == 0 {
		timeout = 60 * time.Second // 60 秒超时
	}

	return &HeartbeatMonitor{
		store:         store,
		logger:        logger,
		metrics:       m,
		checkInterval: checkInterval,
		timeout:       timeout,
		stopCh:        make(chan struct{}),
	}
}

// Start 启动心跳监控
func (m *HeartbeatMonitor) Start(ctx context.Context) {
	m.logger.Info("starting heartbeat monitor",
		zap.Duration("check_interval", m.checkInterval),
		zap.Duration("timeout", m.timeout))

	m.wg.Add(1)
	go m.run(ctx)
}

// Stop 停止心跳监控
func (m *HeartbeatMonitor) Stop() {
	m.logger.Info("stopping heartbeat monitor")
	close(m.stopCh)
	m.wg.Wait()
}

// run 执行心跳监控循环
func (m *HeartbeatMonitor) run(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkHeartbeats(ctx)
		case <-m.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkHeartbeats 检查所有在线 Agent 的心跳
func (m *HeartbeatMonitor) checkHeartbeats(ctx context.Context) {
	// 查询心跳超时的 Agent
	staleAgents, err := m.store.ListStaleAgents(ctx, m.timeout)
	if err != nil {
		m.logger.Error("failed to list stale agents", zap.Error(err))
		return
	}

	// 更新 metrics: 陈旧 Agent 数量
	if m.metrics != nil {
		m.metrics.AgentStaleCount.Set(float64(len(staleAgents)))
	}

	if len(staleAgents) == 0 {
		return
	}

	m.logger.Info("found stale agents", zap.Int("count", len(staleAgents)))

	// 处理每个超时的 Agent
	for _, agent := range staleAgents {
		if err := m.handleHeartbeatTimeout(ctx, agent); err != nil {
			m.logger.Error("failed to handle heartbeat timeout",
				zap.String("agent_id", agent.ID),
				zap.Error(err))
		}
	}
}

// handleHeartbeatTimeout 处理心跳超时的 Agent
func (m *HeartbeatMonitor) handleHeartbeatTimeout(ctx context.Context, agent *model.Agent) error {
	m.logger.Warn("agent heartbeat timeout",
		zap.String("agent_id", agent.ID),
		zap.String("agent_name", agent.Name),
		zap.Time("last_seen_at", *agent.LastSeenAt))

	// 标记 Agent 为离线
	if err := m.store.UpdateAgentStatus(ctx, agent.ID, model.StatusOffline); err != nil {
		return err
	}

	// 更新 metrics: 状态变更
	if m.metrics != nil {
		m.metrics.AgentStatusChanges.WithLabelValues(string(model.StatusOnline), string(model.StatusOffline)).Inc()
		m.metrics.AgentsByStatus.WithLabelValues(string(model.StatusOffline)).Inc()
		m.metrics.AgentsByStatus.WithLabelValues(string(model.StatusOnline)).Dec()
	}

	// 设置断开原因
	reason := "heartbeat timeout"
	if err := m.store.SetAgentDisconnectReason(ctx, agent.ID, reason); err != nil {
		m.logger.Error("failed to set disconnect reason",
			zap.String("agent_id", agent.ID),
			zap.Error(err))
	}

	// 更新活跃的连接历史记录
	activeHistory, err := m.store.GetActiveConnectionHistory(ctx, agent.ID)
	if err != nil {
		m.logger.Error("failed to get active connection history",
			zap.String("agent_id", agent.ID),
			zap.Error(err))
		return nil
	}

	if activeHistory != nil {
		now := time.Now()
		activeHistory.DisconnectedAt = &now
		activeHistory.DisconnectReason = reason
		if err := m.store.UpdateConnectionHistory(ctx, activeHistory); err != nil {
			m.logger.Error("failed to update connection history",
				zap.String("agent_id", agent.ID),
				zap.Error(err))
		}

		// 更新 metrics: 记录连接时长
		if m.metrics != nil && !activeHistory.ConnectedAt.IsZero() {
			duration := now.Sub(activeHistory.ConnectedAt).Seconds()
			m.metrics.AgentConnectionDuration.WithLabelValues(agent.ID).Observe(duration)
		}
	}

	return nil
}
