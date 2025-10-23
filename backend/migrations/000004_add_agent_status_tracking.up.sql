-- 为 agents 表添加状态跟踪字段
ALTER TABLE agents ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'offline';
ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMP;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_connected_at TIMESTAMP;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_disconnected_at TIMESTAMP;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS disconnect_reason TEXT;

-- 添加索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_last_seen_at ON agents(last_seen_at);

-- 创建 agent_connection_history 表
CREATE TABLE IF NOT EXISTS agent_connection_history (
    id BIGSERIAL PRIMARY KEY,
    agent_id VARCHAR(255) NOT NULL,
    connected_at TIMESTAMP NOT NULL,
    disconnected_at TIMESTAMP,
    duration_seconds INTEGER,
    disconnect_reason TEXT,
    remote_addr VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_agent_connection_history_agent
        FOREIGN KEY (agent_id)
        REFERENCES agents(id)
        ON DELETE CASCADE
);

-- 为 agent_connection_history 添加索引
CREATE INDEX IF NOT EXISTS idx_agent_connection_history_agent_id ON agent_connection_history(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_connection_history_connected_at ON agent_connection_history(connected_at);
CREATE INDEX IF NOT EXISTS idx_agent_connection_history_disconnected_at ON agent_connection_history(disconnected_at);

-- 更新现有 agents 的状态为 offline
UPDATE agents SET status = 'offline' WHERE status IS NULL;
