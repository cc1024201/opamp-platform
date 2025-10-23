package model

import (
	"time"
)

// AgentStatus 表示 Agent 的连接状态
type AgentStatus string

const (
	StatusOnline  AgentStatus = "online"  // 在线
	StatusOffline AgentStatus = "offline" // 离线
	StatusError   AgentStatus = "error"   // 错误状态
)

// String 返回状态的字符串表示
func (s AgentStatus) String() string {
	return string(s)
}

// IsValid 检查状态是否有效
func (s AgentStatus) IsValid() bool {
	switch s {
	case StatusOnline, StatusOffline, StatusError:
		return true
	default:
		return false
	}
}

// Agent 代表一个被管理的遥测代理
type Agent struct {
	// 基础信息
	ID           string    `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"index"`
	Type         string    `json:"type"`         // 操作系统类型: linux, windows, darwin
	Architecture string    `json:"architecture"` // CPU 架构: amd64, arm64
	Hostname     string    `json:"hostname" gorm:"index"`
	Version      string    `json:"version"` // Agent 版本

	// 连接状态
	Status              AgentStatus `json:"status" gorm:"type:varchar(20);default:offline;index"`
	LastSeenAt          *time.Time  `json:"last_seen_at,omitempty" gorm:"index"`
	LastConnectedAt     *time.Time  `json:"last_connected_at,omitempty"`
	LastDisconnectedAt  *time.Time  `json:"last_disconnected_at,omitempty"`
	DisconnectReason    string      `json:"disconnect_reason,omitempty"`

	// 兼容性字段 (将来可以移除)
	ConnectedAt    *time.Time  `json:"connected_at,omitempty" gorm:"-"`
	DisconnectedAt *time.Time  `json:"disconnected_at,omitempty" gorm:"-"`

	// 标签 (用于配置匹配)
	Labels Labels `json:"labels" gorm:"serializer:json"`

	// 当前配置
	ConfigurationName string `json:"configuration_name,omitempty"` // 关联的配置名称

	// OpAMP 协议相关
	Protocol       string `json:"protocol"`        // 使用的协议: opamp
	OpAMPState     []byte `json:"-" gorm:"type:bytea"` // OpAMP 状态 (序列化的 protobuf)
	SequenceNumber uint64 `json:"sequence_number"` // OpAMP 消息序列号

	// 元数据
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Agent) TableName() string {
	return "agents"
}

// Labels 表示 Agent 的标签集合
type Labels map[string]string

// Matches 检查标签是否匹配选择器
func (l Labels) Matches(selector map[string]string) bool {
	if len(selector) == 0 {
		return true
	}

	for key, value := range selector {
		labelValue, exists := l[key]
		if !exists || labelValue != value {
			return false
		}
	}

	return true
}

// Merge 合并标签
func (l Labels) Merge(other Labels) Labels {
	result := make(Labels, len(l)+len(other))
	for k, v := range l {
		result[k] = v
	}
	for k, v := range other {
		result[k] = v
	}
	return result
}

// AgentUpdate 表示 Agent 需要接收的更新
type AgentUpdate struct {
	Labels        *Labels        `json:"labels,omitempty"`
	Configuration *Configuration `json:"configuration,omitempty"`
}
