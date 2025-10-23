package model

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Configuration 表示一个遥测配置
type Configuration struct {
	// 基础信息
	Name        string `json:"name" gorm:"primaryKey"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`

	// 配置内容
	ContentType string `json:"content_type"` // yaml, json
	RawConfig   string `json:"raw_config" gorm:"type:text"`    // 原始配置内容 (YAML/JSON)
	ConfigHash  string `json:"config_hash"` // 配置内容的 SHA256 哈希

	// 版本管理
	Version        int        `json:"version" gorm:"default:1"` // 配置版本号
	LastAppliedAt  *time.Time `json:"last_applied_at,omitempty"` // 最后应用时间

	// 选择器 (决定哪些 Agent 使用此配置)
	Selector map[string]string `json:"selector" gorm:"serializer:json"` // 标签选择器

	// 平台配置 (用于组合式配置)
	Platform *PlatformConfig `json:"platform,omitempty" gorm:"serializer:json"`

	// 元数据
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Configuration) TableName() string {
	return "configurations"
}

// PlatformConfig 表示平台配置 (组合式配置)
type PlatformConfig struct {
	Sources      []ResourceReference `json:"sources,omitempty"`
	Processors   []ResourceReference `json:"processors,omitempty"`
	Destinations []ResourceReference `json:"destinations,omitempty"`
}

// ResourceReference 表示对资源的引用
type ResourceReference struct {
	Name       string                 `json:"name"`       // 资源实例名称
	Type       string                 `json:"type"`       // 资源类型
	Parameters map[string]interface{} `json:"parameters"` // 参数值
}

// UpdateHash 更新配置哈希
func (c *Configuration) UpdateHash() {
	hash := sha256.Sum256([]byte(c.RawConfig))
	c.ConfigHash = hex.EncodeToString(hash[:])
}

// MatchesAgent 检查配置是否匹配 Agent
func (c *Configuration) MatchesAgent(agent *Agent) bool {
	if len(c.Selector) == 0 {
		return false // 无选择器的配置不匹配任何 Agent
	}
	return agent.Labels.Matches(c.Selector)
}

// Source 表示数据源定义
type Source struct {
	Name        string                 `json:"name" gorm:"primaryKey"`
	Type        string                 `json:"type" gorm:"index"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Source) TableName() string {
	return "sources"
}

// Destination 表示目标定义
type Destination struct {
	Name        string                 `json:"name" gorm:"primaryKey"`
	Type        string                 `json:"type" gorm:"index"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Destination) TableName() string {
	return "destinations"
}

// Processor 表示处理器定义
type Processor struct {
	Name        string                 `json:"name" gorm:"primaryKey"`
	Type        string                 `json:"type" gorm:"index"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Processor) TableName() string {
	return "processors"
}
