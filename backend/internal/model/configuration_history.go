package model

import "time"

// ConfigurationHistory 表示配置的历史版本
type ConfigurationHistory struct {
	ID                uint              `json:"id" gorm:"primaryKey"`
	ConfigurationName string            `json:"configuration_name" gorm:"index;not null"`
	Version           int               `json:"version" gorm:"not null"`
	ContentType       string            `json:"content_type" gorm:"default:yaml"`
	RawConfig         string            `json:"raw_config" gorm:"type:text;not null"`
	ConfigHash        string            `json:"config_hash" gorm:"not null"`
	Selector          map[string]string `json:"selector" gorm:"serializer:json"`
	Platform          *PlatformConfig   `json:"platform,omitempty" gorm:"serializer:json"`
	ChangeDescription string            `json:"change_description" gorm:"type:text"`
	CreatedBy         string            `json:"created_by"`
	CreatedAt         time.Time         `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (ConfigurationHistory) TableName() string {
	return "configuration_history"
}

// ApplyStatus 表示配置应用的状态
type ApplyStatus string

const (
	ApplyStatusPending  ApplyStatus = "pending"  // 待应用
	ApplyStatusApplying ApplyStatus = "applying" // 应用中
	ApplyStatusApplied  ApplyStatus = "applied"  // 已应用
	ApplyStatusFailed   ApplyStatus = "failed"   // 失败
)

// ConfigurationApplyHistory 表示配置应用到 Agent 的历史记录
type ConfigurationApplyHistory struct {
	ID                uint        `json:"id" gorm:"primaryKey"`
	AgentID           string      `json:"agent_id" gorm:"index;not null"`
	ConfigurationName string      `json:"configuration_name" gorm:"index;not null"`
	ConfigHash        string      `json:"config_hash" gorm:"not null"`
	Status            ApplyStatus `json:"status" gorm:"index;default:pending"`
	ErrorMessage      string      `json:"error_message,omitempty" gorm:"type:text"`
	AppliedAt         *time.Time  `json:"applied_at,omitempty"`
	CreatedAt         time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (ConfigurationApplyHistory) TableName() string {
	return "configuration_apply_history"
}

// IsTerminal 检查状态是否为终态
func (s ApplyStatus) IsTerminal() bool {
	return s == ApplyStatusApplied || s == ApplyStatusFailed
}
