package model

import (
	"time"
)

// AgentConnectionHistory 记录 Agent 的连接历史
type AgentConnectionHistory struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	AgentID   string `json:"agent_id" gorm:"type:varchar(255);not null;index"`

	// 连接信息
	ConnectedAt     time.Time  `json:"connected_at" gorm:"not null;index"`
	DisconnectedAt  *time.Time `json:"disconnected_at,omitempty" gorm:"index"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`

	// 断开原因和元数据
	DisconnectReason string `json:"disconnect_reason,omitempty"`
	RemoteAddr       string `json:"remote_addr,omitempty" gorm:"type:varchar(255)"`

	// 时间戳
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (AgentConnectionHistory) TableName() string {
	return "agent_connection_history"
}

// CalculateDuration 计算并更新连接时长
func (h *AgentConnectionHistory) CalculateDuration() {
	if h.DisconnectedAt != nil {
		duration := int(h.DisconnectedAt.Sub(h.ConnectedAt).Seconds())
		h.DurationSeconds = &duration
	}
}
