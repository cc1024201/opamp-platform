-- 删除索引
DROP INDEX IF EXISTS idx_packages_created_at;
DROP INDEX IF EXISTS idx_packages_is_active;
DROP INDEX IF EXISTS idx_packages_version;
DROP INDEX IF EXISTS idx_packages_platform;
DROP INDEX IF EXISTS idx_package_unique;

-- 删除表
DROP TABLE IF EXISTS packages;
