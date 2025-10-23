package postgres

import (
	"context"
	"time"

	"github.com/cc1024201/opamp-platform/internal/model"
)

// CreateConnectionHistory 创建连接历史记录
func (s *Store) CreateConnectionHistory(ctx context.Context, history *model.AgentConnectionHistory) error {
	return s.db.WithContext(ctx).Create(history).Error
}

// UpdateConnectionHistory 更新连接历史记录
func (s *Store) UpdateConnectionHistory(ctx context.Context, history *model.AgentConnectionHistory) error {
	// 计算连接时长
	history.CalculateDuration()
	return s.db.WithContext(ctx).Save(history).Error
}

// GetConnectionHistory 获取指定的连接历史记录
func (s *Store) GetConnectionHistory(ctx context.Context, id uint) (*model.AgentConnectionHistory, error) {
	var history model.AgentConnectionHistory
	result := s.db.WithContext(ctx).Where("id = ?", id).First(&history)
	if result.Error != nil {
		return nil, result.Error
	}
	return &history, nil
}

// ListConnectionHistoryByAgent 列出指定 Agent 的连接历史
func (s *Store) ListConnectionHistoryByAgent(ctx context.Context, agentID string, limit, offset int) ([]*model.AgentConnectionHistory, int64, error) {
	var histories []*model.AgentConnectionHistory
	var total int64

	// 计算总数
	if err := s.db.WithContext(ctx).Model(&model.AgentConnectionHistory{}).
		Where("agent_id = ?", agentID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := s.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("connected_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// GetActiveConnectionHistory 获取 Agent 当前活跃的连接历史记录
func (s *Store) GetActiveConnectionHistory(ctx context.Context, agentID string) (*model.AgentConnectionHistory, error) {
	var history model.AgentConnectionHistory
	result := s.db.WithContext(ctx).
		Where("agent_id = ? AND disconnected_at IS NULL", agentID).
		Order("connected_at DESC").
		First(&history)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			return nil, nil
		}
		return nil, result.Error
	}
	return &history, nil
}

// UpdateAgentStatus 更新 Agent 状态
func (s *Store) UpdateAgentStatus(ctx context.Context, agentID string, status model.AgentStatus) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"last_seen_at": now,
	}

	if status == model.StatusOnline {
		updates["last_connected_at"] = now
		updates["disconnect_reason"] = "" // 清空断开原因
	} else if status == model.StatusOffline {
		updates["last_disconnected_at"] = now
	}

	return s.db.WithContext(ctx).
		Model(&model.Agent{}).
		Where("id = ?", agentID).
		Updates(updates).Error
}

// UpdateAgentLastSeen 更新 Agent 最后心跳时间
func (s *Store) UpdateAgentLastSeen(ctx context.Context, agentID string) error {
	return s.db.WithContext(ctx).
		Model(&model.Agent{}).
		Where("id = ?", agentID).
		Update("last_seen_at", time.Now()).Error
}

// SetAgentDisconnectReason 设置 Agent 断开原因
func (s *Store) SetAgentDisconnectReason(ctx context.Context, agentID string, reason string) error {
	return s.db.WithContext(ctx).
		Model(&model.Agent{}).
		Where("id = ?", agentID).
		Update("disconnect_reason", reason).Error
}

// ListOnlineAgents 列出所有在线的 Agent
func (s *Store) ListOnlineAgents(ctx context.Context) ([]*model.Agent, error) {
	var agents []*model.Agent
	err := s.db.WithContext(ctx).
		Where("status = ?", model.StatusOnline).
		Find(&agents).Error
	return agents, err
}

// ListOfflineAgents 列出所有离线的 Agent
func (s *Store) ListOfflineAgents(ctx context.Context, limit, offset int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var total int64

	// 计算总数
	if err := s.db.WithContext(ctx).Model(&model.Agent{}).
		Where("status = ?", model.StatusOffline).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := s.db.WithContext(ctx).
		Where("status = ?", model.StatusOffline).
		Order("last_disconnected_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&agents).Error; err != nil {
		return nil, 0, err
	}

	return agents, total, nil
}

// ListStaleAgents 列出心跳超时的 Agent (status 为 online 但 last_seen_at 超过指定时间)
func (s *Store) ListStaleAgents(ctx context.Context, timeout time.Duration) ([]*model.Agent, error) {
	var agents []*model.Agent
	threshold := time.Now().Add(-timeout)

	err := s.db.WithContext(ctx).
		Where("status = ? AND last_seen_at < ?", model.StatusOnline, threshold).
		Find(&agents).Error

	return agents, err
}
