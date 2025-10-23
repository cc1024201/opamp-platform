-- 删除 agent_connection_history 表
DROP TABLE IF EXISTS agent_connection_history;

-- 删除 agents 表的索引
DROP INDEX IF EXISTS idx_agents_status;
DROP INDEX IF EXISTS idx_agents_last_seen_at;

-- 删除 agents 表的新增字段
ALTER TABLE agents DROP COLUMN IF EXISTS status;
ALTER TABLE agents DROP COLUMN IF EXISTS last_seen_at;
ALTER TABLE agents DROP COLUMN IF EXISTS last_connected_at;
ALTER TABLE agents DROP COLUMN IF EXISTS last_disconnected_at;
ALTER TABLE agents DROP COLUMN IF EXISTS disconnect_reason;
