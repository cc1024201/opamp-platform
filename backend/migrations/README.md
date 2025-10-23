# 数据库迁移指南

本目录包含所有数据库迁移文件,使用 [golang-migrate](https://github.com/golang-migrate/migrate) 工具管理数据库 Schema 变更。

## 📚 目录

- [为什么使用迁移](#为什么使用迁移)
- [安装工具](#安装工具)
- [迁移文件命名](#迁移文件命名)
- [常用命令](#常用命令)
- [最佳实践](#最佳实践)
- [故障排查](#故障排查)

---

## 为什么使用迁移

数据库迁移工具提供:

1. **版本控制**: 将数据库 Schema 变更纳入版本控制
2. **可重复性**: 在不同环境(开发、测试、生产)保持一致的数据库结构
3. **可回滚**: 支持迁移的向上(up)和向下(down)
4. **团队协作**: 多人协作时避免数据库冲突
5. **CI/CD 集成**: 自动化部署时自动应用数据库变更

**对比 GORM AutoMigrate**:
- ✅ AutoMigrate: 开发阶段快速迭代
- ✅ Migrate: 生产环境、版本控制、可回滚

---

## 安装工具

### 方式一:使用 Go install (推荐)

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

安装后工具位于 `~/go/bin/migrate`

### 方式二:使用包管理器

**macOS (Homebrew)**:
```bash
brew install golang-migrate
```

**Ubuntu/Debian**:
```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

**验证安装**:
```bash
migrate -version
# 或
~/go/bin/migrate -version
```

---

## 迁移文件命名

迁移文件采用序列号命名:

```
{version}_{description}.{up|down}.sql
```

**示例**:
- `000001_initial_schema.up.sql` - 创建初始表结构
- `000001_initial_schema.down.sql` - 回滚初始表结构
- `000002_add_users_table.up.sql` - 添加用户表
- `000002_add_users_table.down.sql` - 删除用户表

**规则**:
- 版本号必须递增(自动生成)
- 每个迁移必须有 `.up.sql` 和 `.down.sql` 两个文件
- 描述使用小写和下划线,简洁明了

---

## 常用命令

项目已在 `Makefile` 中集成了迁移命令,推荐使用 `make` 命令。

### 1. 查看当前迁移版本

```bash
make migrate-version
```

**输出示例**:
```
1  # 当前版本号
```

### 2. 创建新迁移文件

```bash
make migrate-create name=add_roles_column
```

**生成文件**:
- `000002_add_roles_column.up.sql` - 向上迁移(应用变更)
- `000002_add_roles_column.down.sql` - 向下迁移(回滚变更)

**编辑生成的文件**:

`000002_add_roles_column.up.sql`:
```sql
-- Add role column to users table
ALTER TABLE users ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'user';
CREATE INDEX idx_users_role ON users(role);
```

`000002_add_roles_column.down.sql`:
```sql
-- Remove role column from users table
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN role;
```

### 3. 应用所有待处理的迁移

```bash
make migrate-up
```

**等同于**:
```bash
~/go/bin/migrate -path migrations -database "$DB_URL" up
```

### 4. 回滚最后一次迁移

```bash
make migrate-down
```

**注意**: 只回滚一次,不会回滚所有迁移。

### 5. 迁移到指定版本

```bash
make migrate-goto version=3
```

可以向上或向下迁移到任意版本。

### 6. 强制设置迁移版本

```bash
make migrate-force version=1
```

**用途**: 修复 "dirty" 状态(见故障排查)

**⚠️ 警告**: 不会实际执行迁移,只是标记版本号。

### 7. 删除所有表 (危险!)

```bash
make migrate-drop
```

交互式命令,需要输入 `yes` 确认。

**⚠️ 警告**: 不可恢复,仅用于开发环境!

---

## 最佳实践

### 1. 迁移文件应该是幂等的

使用 `IF EXISTS` / `IF NOT EXISTS`:

```sql
-- ✅ 好的做法
CREATE TABLE IF NOT EXISTS users (...);
DROP TABLE IF EXISTS temp_table;

-- ❌ 避免
CREATE TABLE users (...);  -- 第二次运行会失败
DROP TABLE temp_table;     -- 表不存在时会失败
```

### 2. 每个迁移只做一件事

```sql
-- ✅ 好的做法
-- Migration 1: Add users table
-- Migration 2: Add roles column
-- Migration 3: Add indexes

-- ❌ 避免
-- Migration 1: Add 10 tables, 20 columns, 50 indexes
```

### 3. 向下迁移必须完全撤销向上迁移

```sql
-- up.sql
ALTER TABLE users ADD COLUMN email VARCHAR(255);

-- down.sql
ALTER TABLE users DROP COLUMN email;  -- ✅ 完全撤销
```

### 4. 不要修改已应用的迁移文件

一旦迁移文件被应用到任何环境(开发、测试、生产),就不应该修改它。

如果需要变更,创建新的迁移文件。

### 5. 数据迁移要小心

对于大表的数据变更:
- 考虑分批处理
- 添加超时保护
- 在非高峰期执行

```sql
-- 谨慎处理大表数据迁移
UPDATE users SET role = 'user' WHERE role IS NULL AND id < 10000;
-- 分批执行,避免锁表过久
```

### 6. 测试向下迁移

```bash
make migrate-up      # 应用迁移
make migrate-down    # 立即测试回滚
make migrate-up      # 再次应用,确保可重复
```

---

## 故障排查

### 问题 1: "Dirty database version"

**现象**:
```
error: Dirty database version 1. Fix and force version.
```

**原因**: 迁移过程中失败,数据库处于不一致状态。

**解决方案**:

1. **检查数据库实际状态**:
   ```sql
   -- 连接数据库
   psql -U opamp -d opamp_platform

   -- 查看 schema_migrations 表
   SELECT * FROM schema_migrations;
   ```

2. **手动修复数据库**:
   - 如果迁移已部分应用,手动完成或回滚
   - 删除有问题的表/列

3. **强制设置版本**:
   ```bash
   # 如果数据库实际是版本 1
   make migrate-force version=1

   # 然后重新尝试
   make migrate-up
   ```

### 问题 2: "relation already exists"

**现象**:
```
error: pq: relation "users" already exists
```

**原因**: 表已经存在(可能通过 GORM AutoMigrate 创建)

**解决方案**:

1. **如果是初次使用迁移**:
   ```bash
   # 强制设置为当前版本(不执行迁移)
   make migrate-force version=1
   ```

2. **如果需要清理重来**:
   ```bash
   make migrate-drop     # 删除所有表
   make migrate-up       # 重新应用迁移
   ```

### 问题 3: "no migration"

**现象**:
```
error: no migration
```

**原因**: 数据库没有运行过任何迁移。

**解决方案**:
```bash
make migrate-up  # 应用第一次迁移
```

### 问题 4: 连接数据库失败

**现象**:
```
error: dial tcp 127.0.0.1:5432: connect: connection refused
```

**解决方案**:
```bash
# 启动 Docker 服务
make docker-up

# 等待几秒,然后重试
make migrate-up
```

### 问题 5: 迁移卡住不动

**可能原因**: 数据库表被锁定

**解决方案**:
```sql
-- 查看锁表情况
SELECT * FROM pg_locks WHERE NOT granted;

-- 杀死锁表的进程(谨慎!)
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'active' AND pid != pg_backend_pid();
```

---

## 开发工作流

### 场景 1: 添加新表

```bash
# 1. 创建迁移文件
make migrate-create name=add_audit_logs_table

# 2. 编辑 up.sql
# migrations/000002_add_audit_logs_table.up.sql
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    action VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

# 3. 编辑 down.sql
# migrations/000002_add_audit_logs_table.down.sql
DROP TABLE IF EXISTS audit_logs;

# 4. 应用迁移
make migrate-up

# 5. 测试回滚
make migrate-down
make migrate-up
```

### 场景 2: 修改现有表

```bash
# 1. 创建迁移
make migrate-create name=add_email_to_users

# 2. up.sql
ALTER TABLE users ADD COLUMN email VARCHAR(255);
CREATE UNIQUE INDEX idx_users_email ON users(email);

# 3. down.sql
DROP INDEX IF EXISTS idx_users_email;
ALTER TABLE users DROP COLUMN email;

# 4. 应用
make migrate-up
```

### 场景 3: 数据迁移

```bash
# 1. 创建迁移
make migrate-create name=migrate_user_roles

# 2. up.sql - 数据转换
UPDATE users SET role = 'admin' WHERE username = 'admin';
UPDATE users SET role = 'user' WHERE role IS NULL;

# 3. down.sql - 回滚数据(如果可能)
UPDATE users SET role = NULL WHERE role IN ('admin', 'user');

# 4. 谨慎应用(先备份!)
make migrate-up
```

---

## 环境配置

迁移使用的数据库连接信息在 `Makefile` 中定义:

```makefile
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= opamp
DB_PASSWORD ?= opamp123
DB_NAME ?= opamp_platform
```

**覆盖默认值**:
```bash
# 临时覆盖
DB_HOST=192.168.31.46 make migrate-up

# 或设置环境变量
export DB_HOST=192.168.31.46
make migrate-up
```

---

## 与 GORM 的集成

项目同时支持两种模式:

### 开发阶段(推荐 AutoMigrate)

```go
// internal/store/postgres/store.go
func NewStore(...) {
    // 自动创建/更新表结构
    db.AutoMigrate(&model.Agent{}, &model.Configuration{}, &model.User{})
}
```

**优点**: 快速迭代,自动同步模型变更

### 生产环境(推荐 golang-migrate)

1. **禁用 AutoMigrate**:
   ```go
   // 生产环境不使用 AutoMigrate
   // db.AutoMigrate(...)
   ```

2. **使用迁移文件**:
   ```bash
   make migrate-up
   ```

**优点**: 可控、可回滚、可审计

### 混合使用

可以保留 AutoMigrate,但使用 migrate 管理重要变更:
- 日常开发: AutoMigrate
- 重要变更(添加索引、数据迁移): migrate

---

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Database Migration

on:
  push:
    branches: [main]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install migrate
        run: |
          curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
          sudo mv migrate /usr/local/bin/

      - name: Run migrations
        env:
          DB_URL: ${{ secrets.DATABASE_URL }}
        run: |
          cd backend
          migrate -path migrations -database "$DB_URL" up
```

---

## 参考资料

- [golang-migrate 官方文档](https://github.com/golang-migrate/migrate)
- [PostgreSQL 数据库迁移最佳实践](https://www.postgresql.org/docs/current/ddl-alter.html)
- [数据库版本控制](https://www.liquibase.org/get-started/database-version-control)

---

**最后更新**: 2025-10-23
**维护者**: OpAMP Platform 开发团队
