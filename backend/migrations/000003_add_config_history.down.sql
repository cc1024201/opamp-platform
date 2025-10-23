-- 回滚配置历史功能

-- 删除索引
DROP INDEX IF EXISTS idx_apply_history_created_at;
DROP INDEX IF EXISTS idx_apply_history_status;
DROP INDEX IF EXISTS idx_apply_history_config;
DROP INDEX IF EXISTS idx_apply_history_agent;

DROP INDEX IF EXISTS idx_config_history_version;
DROP INDEX IF EXISTS idx_config_history_created_at;
DROP INDEX IF EXISTS idx_config_history_name;

-- 删除表
DROP TABLE IF EXISTS configuration_apply_history;
DROP TABLE IF EXISTS configuration_history;

-- 删除 configurations 表的新字段
ALTER TABLE configurations DROP COLUMN IF EXISTS last_applied_at;
ALTER TABLE configurations DROP COLUMN IF EXISTS version;
