-- 创建 packages 表
CREATE TABLE IF NOT EXISTS packages (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    arch VARCHAR(50) NOT NULL,
    file_size BIGINT,
    checksum VARCHAR(64),
    storage_path VARCHAR(500),
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建唯一索引 (name, version, platform, arch 组合唯一)
CREATE UNIQUE INDEX idx_package_unique
    ON packages(name, version, platform, arch);

-- 创建查询优化索引
CREATE INDEX idx_packages_platform ON packages(platform);
CREATE INDEX idx_packages_version ON packages(version);
CREATE INDEX idx_packages_is_active ON packages(is_active);
CREATE INDEX idx_packages_created_at ON packages(created_at DESC);

-- 添加注释
COMMENT ON TABLE packages IS 'Agent 软件包管理表';
COMMENT ON COLUMN packages.name IS '包名称';
COMMENT ON COLUMN packages.version IS '版本号';
COMMENT ON COLUMN packages.platform IS '平台 (linux/windows/darwin)';
COMMENT ON COLUMN packages.arch IS '架构 (amd64/arm64/386)';
COMMENT ON COLUMN packages.file_size IS '文件大小(字节)';
COMMENT ON COLUMN packages.checksum IS 'SHA256 校验和';
COMMENT ON COLUMN packages.storage_path IS 'MinIO 存储路径';
COMMENT ON COLUMN packages.is_active IS '是否激活';
