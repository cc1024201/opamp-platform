-- 配置历史版本表
CREATE TABLE IF NOT EXISTS configuration_history (
    id BIGSERIAL PRIMARY KEY,
    configuration_name VARCHAR(255) NOT NULL,
    version INTEGER NOT NULL,
    content_type VARCHAR(50) NOT NULL DEFAULT 'yaml',
    raw_config TEXT NOT NULL,
    config_hash VARCHAR(64) NOT NULL,
    selector JSONB,
    platform JSONB,
    change_description TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- 外键约束
    CONSTRAINT fk_configuration
        FOREIGN KEY (configuration_name)
        REFERENCES configurations(name)
        ON DELETE CASCADE,

    -- 唯一约束: 每个配置的版本号必须唯一
    CONSTRAINT unique_config_version
        UNIQUE (configuration_name, version)
);

-- 索引
CREATE INDEX idx_config_history_name ON configuration_history(configuration_name);
CREATE INDEX idx_config_history_created_at ON configuration_history(created_at);
CREATE INDEX idx_config_history_version ON configuration_history(configuration_name, version DESC);

-- 配置应用记录表 (记录配置推送到 Agent 的历史)
CREATE TABLE IF NOT EXISTS configuration_apply_history (
    id BIGSERIAL PRIMARY KEY,
    agent_id VARCHAR(255) NOT NULL,
    configuration_name VARCHAR(255) NOT NULL,
    config_hash VARCHAR(64) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, applying, applied, failed
    error_message TEXT,
    applied_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- 外键约束
    CONSTRAINT fk_agent
        FOREIGN KEY (agent_id)
        REFERENCES agents(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_config_apply
        FOREIGN KEY (configuration_name)
        REFERENCES configurations(name)
        ON DELETE CASCADE
);

-- 索引
CREATE INDEX idx_apply_history_agent ON configuration_apply_history(agent_id);
CREATE INDEX idx_apply_history_config ON configuration_apply_history(configuration_name);
CREATE INDEX idx_apply_history_status ON configuration_apply_history(status);
CREATE INDEX idx_apply_history_created_at ON configuration_apply_history(created_at DESC);

-- 添加版本号字段到 configurations 表
ALTER TABLE configurations ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;
ALTER TABLE configurations ADD COLUMN IF NOT EXISTS last_applied_at TIMESTAMP WITH TIME ZONE;

-- 为配置名称添加注释
COMMENT ON TABLE configuration_history IS '配置历史版本记录表';
COMMENT ON TABLE configuration_apply_history IS '配置应用历史记录表';
COMMENT ON COLUMN configuration_history.version IS '配置版本号,从 1 开始递增';
COMMENT ON COLUMN configuration_apply_history.status IS '应用状态: pending(待应用), applying(应用中), applied(已应用), failed(失败)';
