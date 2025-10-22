# OpAMP Platform

一个基于 [OpenTelemetry OpAMP](https://github.com/open-telemetry/opamp-spec) 协议的现代化Agent管理平台。

## 🎯 项目特性

- ✅ **最新技术栈**: 基于 opamp-go v0.22.0, Go 1.24, PostgreSQL 16
- ✅ **稳定可靠**: 使用企业级数据库和缓存方案
- ✅ **易于扩展**: 清晰的架构设计，模块化开发
- ✅ **生产就绪**: 完整的日志、监控、健康检查

## 📦 技术栈

### 后端
- **语言**: Go 1.24
- **框架**: Gin v1.11
- **OpAMP**: opamp-go v0.22.0 (官方最新版本)
- **数据库**: PostgreSQL 16 + GORM
- **缓存**: Redis 7
- **存储**: MinIO (S3兼容)
- **日志**: Zap v1.27

### 前端 (计划中)
- **框架**: React 18
- **构建**: Vite 5
- **语言**: TypeScript 5
- **UI库**: Material-UI v6

## 🚀 快速开始

### 1. 启动开发环境

```bash
# 克隆项目
cd opamp-platform

# 启动数据库和缓存服务
docker-compose up -d

# 等待服务启动
docker-compose ps

# 查看服务状态
# postgres (端口 5432)
# redis (端口 6379)
# minio (端口 9000/9001)
```

### 2. 编译并运行服务器

```bash
cd backend

# 编译
go build -o bin/opamp-server ./cmd/server

# 运行
./bin/opamp-server
```

服务器将在 http://localhost:8080 启动。

### 3. 访问服务

- **API**: http://localhost:8080/api/v1
- **健康检查**: http://localhost:8080/health
- **OpAMP 端点**: ws://localhost:8080/v1/opamp
- **MinIO 控制台**: http://localhost:9001 (minioadmin/minioadmin123)
- **PostgreSQL**: localhost:5432 (opamp/opamp123/opamp_platform)

### 4. (可选) 启动 pgAdmin

```bash
docker-compose --profile tools up -d

# 访问 pgAdmin: http://localhost:5050
# 登录: admin@opamp.local / admin123
```

## 📚 API 文档

### Agent 管理

```bash
# 列出所有 Agent
curl http://localhost:8080/api/v1/agents

# 获取单个 Agent
curl http://localhost:8080/api/v1/agents/{agent-id}

# 删除 Agent
curl -X DELETE http://localhost:8080/api/v1/agents/{agent-id}
```

### Configuration 管理

```bash
# 列出所有配置
curl http://localhost:8080/api/v1/configurations

# 获取单个配置
curl http://localhost:8080/api/v1/configurations/{name}

# 创建配置
curl -X POST http://localhost:8080/api/v1/configurations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod-config",
    "display_name": "生产环境配置",
    "content_type": "yaml",
    "raw_config": "receivers:\n  otlp:\n    protocols:\n      grpc:\n...",
    "selector": {
      "env": "prod",
      "region": "us-east"
    }
  }'

# 更新配置
curl -X PUT http://localhost:8080/api/v1/configurations/{name} \
  -H "Content-Type: application/json" \
  -d '{ ... }'

# 删除配置
curl -X DELETE http://localhost:8080/api/v1/configurations/{name}
```

## 🏗️ 项目结构

```
opamp-platform/
├── backend/                    # 后端代码
│   ├── cmd/
│   │   └── server/            # 主程序入口
│   │       ├── main.go
│   │       └── handlers.go    # API 处理函数
│   ├── internal/
│   │   ├── model/             # 数据模型
│   │   │   ├── agent.go
│   │   │   └── configuration.go
│   │   ├── opamp/             # OpAMP 服务器实现
│   │   │   ├── server.go
│   │   │   ├── callbacks.go
│   │   │   └── logger.go
│   │   └── store/             # 存储层
│   │       └── postgres/
│   │           └── store.go
│   ├── config.yaml            # 配置文件
│   └── go.mod
│
├── frontend/                   # 前端代码 (TODO)
│   └── src/
│
├── deploy/                     # 部署配置
├── docs/                       # 文档
├── docker-compose.yml          # 开发环境
└── README.md
```

## 🔧 配置

编辑 `backend/config.yaml`:

```yaml
server:
  port: 8080
  mode: debug

opamp:
  endpoint: /v1/opamp
  secret_key: ""  # 留空则不验证

database:
  host: localhost
  port: 5432
  user: opamp
  password: opamp123
  dbname: opamp_platform

redis:
  host: localhost
  port: 6379

minio:
  endpoint: localhost:9000
  access_key: minioadmin
  secret_key: minioadmin123
```

## 📖 核心概念

### Agent

Agent 是被管理的遥测代理实例，包含：
- 唯一标识 (UUID)
- 基本信息 (名称、版本、主机名、架构)
- 连接状态 (Connected/Disconnected/Configuring/Error)
- 标签 (用于配置匹配)

### Configuration

Configuration 定义了 Agent 的遥测配置：
- 配置内容 (YAML 格式)
- 标签选择器 (决定哪些 Agent 使用此配置)
- 配置哈希 (用于变更检测)

### 配置分发流程

```
1. 创建 Configuration → 设置 selector (env=prod)
                    ↓
2. Agent 连接 → 发送自身标签 (env=prod, region=us-east)
                    ↓
3. 服务器匹配 → 找到匹配的 Configuration
                    ↓
4. 下发配置 → 通过 OpAMP 协议发送给 Agent
                    ↓
5. Agent 应用 → 返回应用状态 (成功/失败)
```

## 🧪 测试

### 模拟 Agent 连接

使用 opamp-go 提供的示例 Agent:

```bash
cd /path/to/opamp-go
cd internal/examples/agent

# 编译示例 Agent
go build -o agent-test .

# 连接到服务器（注意：需要禁用TLS）
./agent-test -initial-insecure-connection
```

**测试结果**: ✅ 已验证
- Agent 连接成功
- 自动注册到数据库
- 配置自动分发成功

详细测试报告: [TESTING_REPORT_v1.md](docs/TESTING_REPORT_v1.md)

## 🗺️ Roadmap

### ✅ Phase 1: 基础架构 (已完成)
- [x] 项目初始化
- [x] Docker Compose 开发环境
- [x] OpAMP Server 集成
- [x] PostgreSQL 存储层
- [x] REST API 框架
- [x] Agent 和 Configuration 数据模型
- [x] OpAMP Agent 连接测试
- [x] 配置分发流程验证
- [x] 完整测试报告

### 🚧 Phase 2: 核心功能 (计划中)
- [ ] 前端 UI 初始化
- [ ] Agent 列表和详情页面
- [ ] Configuration 管理界面
- [ ] 实时状态更新

### 📋 Phase 3: 高级功能 (计划中)
- [ ] GraphQL API
- [ ] WebSocket 实时通知
- [ ] Dashboard 仪表盘
- [ ] 告警系统
- [ ] 用户认证和权限

### 🎯 Phase 4: 生产就绪 (计划中)
- [ ] 高可用部署
- [ ] Kubernetes Operator
- [ ] 监控和日志收集
- [ ] 性能优化
- [ ] 完整文档

## 📝 开发笔记

### 关键设计决策

1. **使用最新 opamp-go (v0.22.0)**
   - 相比 bindplane-op 的 v0.2.0,新版 API 更加清晰
   - 回调函数支持 per-connection callbacks
   - 更好的错误处理和日志支持

2. **PostgreSQL 替代 BoltDB**
   - 支持并发访问
   - 更好的查询能力
   - 易于扩展和备份

3. **模块化设计**
   - 清晰的分层架构
   - 易于测试和扩展
   - 符合 Go 最佳实践

## 🤝 贡献

欢迎提交 Issue 和 Pull Request!

## 📄 许可证

Apache License 2.0

---

**当前状态**: 🚧 开发中 - MVP 阶段

**最后更新**: 2025-10-22
