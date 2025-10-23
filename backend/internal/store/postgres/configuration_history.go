package postgres

import (
	"context"

	"github.com/cc1024201/opamp-platform/internal/model"
)

// CreateConfigurationHistory 创建配置历史记录
func (s *Store) CreateConfigurationHistory(ctx context.Context, history *model.ConfigurationHistory) error {
	return s.db.WithContext(ctx).Create(history).Error
}

// GetConfigurationHistory 获取指定版本的配置历史
func (s *Store) GetConfigurationHistory(ctx context.Context, configName string, version int) (*model.ConfigurationHistory, error) {
	var history model.ConfigurationHistory
	err := s.db.WithContext(ctx).
		Where("configuration_name = ? AND version = ?", configName, version).
		First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// ListConfigurationHistory 列出配置的所有历史版本
func (s *Store) ListConfigurationHistory(ctx context.Context, configName string, limit, offset int) ([]*model.ConfigurationHistory, int64, error) {
	var histories []*model.ConfigurationHistory
	var total int64

	query := s.db.WithContext(ctx).
		Where("configuration_name = ?", configName).
		Order("version DESC")

	// 计算总数
	if err := query.Model(&model.ConfigurationHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Limit(limit).Offset(offset).Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// GetLatestConfigurationVersion 获取配置的最新版本号
func (s *Store) GetLatestConfigurationVersion(ctx context.Context, configName string) (int, error) {
	var maxVersion int
	err := s.db.WithContext(ctx).
		Model(&model.ConfigurationHistory{}).
		Where("configuration_name = ?", configName).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	return maxVersion, err
}

// CreateApplyHistory 创建配置应用历史记录
func (s *Store) CreateApplyHistory(ctx context.Context, history *model.ConfigurationApplyHistory) error {
	return s.db.WithContext(ctx).Create(history).Error
}

// UpdateApplyHistory 更新配置应用历史记录
func (s *Store) UpdateApplyHistory(ctx context.Context, history *model.ConfigurationApplyHistory) error {
	return s.db.WithContext(ctx).Save(history).Error
}

// GetApplyHistory 获取指定 ID 的应用历史
func (s *Store) GetApplyHistory(ctx context.Context, id uint) (*model.ConfigurationApplyHistory, error) {
	var history model.ConfigurationApplyHistory
	err := s.db.WithContext(ctx).First(&history, id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// GetLatestApplyHistory 获取 Agent 最新的配置应用记录
func (s *Store) GetLatestApplyHistory(ctx context.Context, agentID, configName string) (*model.ConfigurationApplyHistory, error) {
	var history model.ConfigurationApplyHistory
	err := s.db.WithContext(ctx).
		Where("agent_id = ? AND configuration_name = ?", agentID, configName).
		Order("created_at DESC").
		First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// ListApplyHistoryByAgent 列出 Agent 的所有配置应用历史
func (s *Store) ListApplyHistoryByAgent(ctx context.Context, agentID string, limit, offset int) ([]*model.ConfigurationApplyHistory, int64, error) {
	var histories []*model.ConfigurationApplyHistory
	var total int64

	query := s.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("created_at DESC")

	// 计算总数
	if err := query.Model(&model.ConfigurationApplyHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Limit(limit).Offset(offset).Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// ListApplyHistoryByConfig 列出配置的所有应用历史
func (s *Store) ListApplyHistoryByConfig(ctx context.Context, configName string, limit, offset int) ([]*model.ConfigurationApplyHistory, int64, error) {
	var histories []*model.ConfigurationApplyHistory
	var total int64

	query := s.db.WithContext(ctx).
		Where("configuration_name = ?", configName).
		Order("created_at DESC")

	// 计算总数
	if err := query.Model(&model.ConfigurationApplyHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Limit(limit).Offset(offset).Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// GetPendingApplyHistories 获取所有待应用或应用中的记录
func (s *Store) GetPendingApplyHistories(ctx context.Context) ([]*model.ConfigurationApplyHistory, error) {
	var histories []*model.ConfigurationApplyHistory
	err := s.db.WithContext(ctx).
		Where("status IN ?", []model.ApplyStatus{model.ApplyStatusPending, model.ApplyStatusApplying}).
		Order("created_at ASC").
		Find(&histories).Error
	return histories, err
}
