# OpAMP Platform 开发指南

**最后更新**: 2025-10-22

本文档是项目的开发指南，包含架构设计、技术决策、开发环境配置和常见问题解决方案。

---

## 📋 目录

1. [项目背景](#项目背景)
2. [技术选型](#技术选型)
3. [架构设计](#架构设计)
4. [技术决策 (ADR)](#技术决策-adr)
5. [技术栈](#技术栈)
6. [开发环境配置](#开发环境配置)
7. [常见问题和解决方案](#常见问题和解决方案)
8. [术语表](#术语表)

---

## 🌱 项目背景

### 为什么需要 OpAMP Platform？

在可观测性领域，管理大规模的遥测 Agent（Telemetry Agents）是一个关键挑战：

**主要痛点**：
1. **配置管理复杂** - 数千个 Agent 的配置难以统一管理
2. **版本控制困难** - Agent 版本分散，升级流程不统一
3. **状态监控缺失** - 无法实时了解 Agent 的运行状态
4. **手动操作繁琐** - 配置变更需要逐个 Agent 操作

**OpAMP 协议的价值**：
- **标准化** - OpenTelemetry 官方定义的 Agent 管理协议
- **自动化** - 远程配置分发、版本升级
- **可观测** - Agent 状态实时上报
- **灵活性** - 标签选择器实现细粒度控制

### 项目目标

构建一个基于 OpenTelemetry OpAMP 协议的**现代化、稳定、易扩展**的 Agent 管理平台：

1. **易用性** - 简单的 UI 界面，降低使用门槛
2. **可靠性** - 企业级数据库，高可用部署
3. **可扩展** - 清晰的架构，支持大规模部署
4. **长期发展** - 使用最新技术，避免技术债务

---

## 🔍 技术选型

### 面临的选择

项目启动时，我们研究了两个主要方案：

#### 方案 A: opamp-go (OpenTelemetry 官方)

**位置**: https://github.com/open-telemetry/opamp-go

**特点**：
- ✅ OpenTelemetry 官方维护
- ✅ 持续更新（最后更新: 2025-10-01）
- ✅ 最新 OpAMP 协议支持
- ✅ API 清晰，文档完善
- ❌ 仅是协议库，无完整功能
- ❌ 需要自己实现 UI 和业务逻辑

#### 方案 B: bindplane-op (observIQ)

**位置**: https://github.com/observIQ/bindplane-op

**特点**：
- ✅ 功能完整（Agent 管理、配置管理、UI）
- ✅ UI 美观实用
- ✅ 成熟的架构设计
- ✅ 可以直接使用
- ❌ 停止维护 3 年（最后更新: 2022-07-29）
- ❌ 依赖过时（Go 1.18, opamp-go v0.2.0）
- ❌ 大量技术债务

### 最终决策

**✅ 选择：基于 opamp-go v0.22.0 构建新平台**

**核心理由**：

1. **稳定性优先**
   - 官方维护，持续更新
   - 跟随 OpenTelemetry 生态演进
   - 避免 3 年技术债务

2. **长期发展**
   - 使用最新稳定技术栈
   - API 设计更加清晰
   - 社区活跃，问题能得到解决

3. **学习优秀设计**
   - 深入分析 bindplane-op 架构
   - 提取可复用的设计模式
   - 重新实现，而不是 fork

4. **自主可控**
   - 完全掌握代码库
   - 可以根据需求定制
   - 不依赖停止维护的项目

### 从 bindplane-op 学到的经验

通过深入分析 bindplane-op，我们提取了以下优秀设计：

**前端设计**：
- 3步配置向导
- 双模态编辑器（UI + YAML）
- 实时状态订阅

**后端架构**：
- 分层架构（API → Manager → Store）
- Protocol 接口抽象
- 标签选择器系统
- EventBus 事件通知

**OpAMP 集成**：
- per-connection callbacks
- 线程安全的连接管理
- 配置哈希变更检测

这些经验指导了我们的架构设计。

---

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        前端 UI Layer                         │
│  React 18 + TypeScript + MUI v6 + Vite                      │
└──────────────────┬──────────────────────────────────────────┘
                   │ REST API + GraphQL + WebSocket
┌──────────────────┴──────────────────────────────────────────┐
│                      API Gateway Layer                       │
│  ├─ REST API (Gin)                                          │
│  ├─ GraphQL (gqlgen) [计划]                                 │
│  └─ WebSocket (Gorilla)                                     │
└──────────────────┬──────────────────────────────────────────┘
                   │
┌──────────────────┴──────────────────────────────────────────┐
│                      业务逻辑 Layer                          │
│  ├─ Agent Manager                                           │
│  ├─ Config Manager                                          │
│  └─ Package Manager [计划]                                  │
└──────────────────┬──────────────────────────────────────────┘
                   │
┌──────────────────┴──────────────────────────────────────────┐
│                   OpAMP Protocol Layer                       │
│          opamp-go v0.22.0                                   │
│  ├─ WebSocket Server                                        │
│  └─ Message Handler                                         │
└──────────────────┬──────────────────────────────────────────┘
                   │
┌──────────────────┴──────────────────────────────────────────┐
│                      存储 Layer                              │
│  ├─ PostgreSQL (主数据)                                     │
│  ├─ Redis (缓存/实时状态)                                   │
│  └─ MinIO (Agent 包存储)                                    │
└─────────────────────────────────────────────────────────────┘
```

### 核心设计原则

1. **分层架构**: 清晰的层次划分，职责分离
2. **接口驱动**: 使用接口定义，易于测试和替换
3. **模块化**: 独立的功能模块，低耦合高内聚
4. **可扩展性**: 支持水平扩展，高可用部署

### 数据模型设计

#### Agent 模型
```go
type Agent struct {
    ID              string      // UUID
    Name            string      // 服务名称
    Type            string      // OS 类型
    Architecture    string      // CPU 架构
    Hostname        string      // 主机名
    Version         string      // Agent 版本
    Status          AgentStatus // 连接状态
    ConnectedAt     *time.Time  // 连接时间
    DisconnectedAt  *time.Time  // 断开时间
    Labels          Labels      // 标签 (JSONB)
    ConfigurationName string    // 关联配置
    OpAMPState      []byte      // OpAMP 状态
    SequenceNumber  uint64      // 序列号
}
```

**关键设计**:
- Labels 使用 JSONB 存储 → 灵活的标签系统
- Status 枚举类型 → 类型安全的状态管理
- SequenceNumber → OpAMP 协议消息顺序

#### Configuration 模型
```go
type Configuration struct {
    Name        string              // 配置名称
    DisplayName string              // 显示名称
    Description string              // 描述
    ContentType string              // 内容类型 (yaml/json)
    RawConfig   string              // 原始配置内容
    ConfigHash  string              // SHA256 哈希
    Selector    map[string]string   // 标签选择器 (JSONB)
    Platform    *PlatformConfig     // 平台配置
}
```

**关键设计**:
- Selector 使用 JSONB → 灵活的标签匹配
- ConfigHash 自动计算 → 配置变更检测
- 支持 YAML/JSON 格式

### 配置分发算法

```go
// Configuration.MatchesAgent
func (c *Configuration) MatchesAgent(agent *Agent) bool {
    if len(c.Selector) == 0 {
        return false  // 空选择器不匹配任何 Agent
    }
    return agent.Labels.Matches(c.Selector)
}

// Labels.Matches - 子集匹配算法
func (l Labels) Matches(selector map[string]string) bool {
    for key, value := range selector {
        if l[key] != value {
            return false
        }
    }
    return true
}
```

**匹配逻辑**:
- selector 必须是 agent labels 的子集
- 支持多标签组合匹配
- 空 selector 不匹配任何 Agent

---

## 📝 技术决策 (ADR)

### ADR-001: 使用 PostgreSQL 替代 BoltDB

**日期**: 2025-10-22
**状态**: ✅ 已采纳

#### 背景
bindplane-op 使用 BoltDB (嵌入式数据库)，但存在以下限制：
- 单机部署，无法水平扩展
- 并发性能有限
- 备份和恢复复杂

#### 决策
使用 PostgreSQL 16 作为主数据库

#### 理由
1. **企业级可靠性** - 成熟稳定，生产验证
2. **支持集群和复制** - 高可用部署
3. **强大的查询能力** - 支持复杂查询和索引
4. **JSONB 支持** - 灵活的 schema
5. **丰富的工具生态** - pgAdmin, pg_dump 等

#### 影响
- ✅ 支持高可用部署
- ✅ 更好的并发性能
- ✅ 标准 SQL 查询
- ❌ 需要独立部署数据库
- ❌ 增加运维复杂度

---

### ADR-002: 使用最新 opamp-go v0.22.0

**日期**: 2025-10-22
**状态**: ✅ 已采纳

#### 背景
bindplane-op 使用 opamp-go v0.2.0 (2022)，官方已更新到 v0.22.0

#### 决策
使用最新版本 opamp-go v0.22.0

#### 理由
1. **官方持续维护** - OpenTelemetry 官方项目
2. **API 更加清晰** - per-connection callbacks
3. **支持最新协议特性** - 跟随 OpAMP 规范更新
4. **更好的错误处理** - 改进的错误类型和日志
5. **长期技术支持** - 避免技术债务

#### API 变化

| 特性 | v0.2.0 | v0.22.0 | 变化 |
|------|--------|---------|------|
| **Callbacks** | `CallbacksStruct` | `Callbacks` + `ConnectionCallbacks` | 分离为连接级回调 |
| **OnConnecting** | 返回简单响应 | 返回 `ConnectionCallbacks` | 更灵活的回调管理 |
| **Logger** | 2参数 `Debugf` | 3参数 `Debugf(ctx, ...)` | 支持 context |

#### 影响
- ✅ 长期技术支持
- ✅ 最新协议特性
- ✅ 更清晰的 API 设计
- ❌ 需要适配新 API
- ❌ 无法直接参考 bindplane-op 代码

---

## 🛠️ 技术栈

### 后端

| 组件 | 版本 | 选择理由 |
|------|------|----------|
| **Go** | 1.24 | 最新稳定版，性能优秀 |
| **opamp-go** | v0.22.0 | 官方最新版，持续更新 |
| **Gin** | v1.11.0 | 高性能 Web 框架 |
| **GORM** | v1.31.0 | 成熟 ORM，易用 |
| **PostgreSQL** | 16 | 企业级数据库 |
| **Redis** | 7 | 高性能缓存 |
| **MinIO** | latest | S3 兼容存储 |
| **Zap** | v1.27.0 | 高性能日志库 |
| **Viper** | v1.21.0 | 配置管理 |

### 前端 (计划中)

| 组件 | 版本 | 选择理由 |
|------|------|----------|
| **React** | 18 | 最新稳定版 |
| **TypeScript** | 5 | 类型安全 |
| **Vite** | 5 | 极速构建 |
| **MUI** | v6 | 企业级组件库 |

### 与 bindplane-op 的对比

| 组件 | bindplane-op | 新平台 | 提升 |
|------|--------------|--------|------|
| Go | 1.18 | 1.24 | +6 个版本 |
| opamp-go | v0.2.0 (2022) | v0.22.0 (2025) | +20 个版本 |
| 数据库 | BoltDB | PostgreSQL | 单机→集群 |
| 缓存 | 无 | Redis | 新增 |

---

## 🚀 开发环境配置

### 前置要求

- Go 1.24+
- Docker & Docker Compose
- Git

### 1. 克隆项目

```bash
git clone https://github.com/cc1024201/opamp-platform.git
cd opamp-platform
```

### 2. 启动基础服务

```bash
# 启动 PostgreSQL, Redis, MinIO
docker-compose up -d

# 验证服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

**服务端口**:
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- MinIO API: `localhost:9000`
- MinIO Console: `localhost:9001` (minioadmin/minioadmin123)

### 3. 配置后端

```bash
cd backend

# 安装依赖
go mod download

# 配置文件
cp config.yaml.example config.yaml
# 根据需要修改 config.yaml

# 编译
go build -o bin/opamp-server ./cmd/server

# 运行
./bin/opamp-server
```

**配置说明** (`config.yaml`):
```yaml
server:
  port: 8080
  mode: debug  # debug | release

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

### 4. CI/CD 配置

#### GitHub Actions

项目使用 GitHub Actions 进行 CI/CD，配置文件：`.github/workflows/test.yml`

**包含的 Job**:
1. **Test** - 运行单元测试，生成覆盖率报告
2. **Lint** - golangci-lint 代码质量检查
3. **Build** - 编译验证

**触发条件**:
- Push 到 `main` 或 `develop` 分支
- Pull Request 到 `main` 或 `develop` 分支

#### Codecov

测试覆盖率自动上传到 Codecov：
```yaml
- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out
```

### 5. 本地开发工具

#### 运行测试
```bash
cd backend

# 运行所有测试
go test ./internal/... -v

# 生成覆盖率报告
go test ./internal/... -cover -coverprofile=coverage.out

# 查看覆盖率
go tool cover -func=coverage.out

# HTML 覆盖率报告
go tool cover -html=coverage.out -o coverage.html
```

#### 代码检查
```bash
# 安装 golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行 linter
golangci-lint run

# 自动修复
golangci-lint run --fix
```

---

## 🐛 常见问题和解决方案

### 问题 1: JSONB 字段扫描错误

**现象**:
```
sql: Scan error on column index 6, name "selector":
unsupported Scan, storing driver.Value type []uint8 into type *map[string]string
```

**原因**:
使用了错误的 GORM 标签 `gorm:"type:jsonb"`

**解决方案**:
```go
// ❌ 错误
Selector map[string]string `json:"selector" gorm:"type:jsonb"`

// ✅ 正确
Selector map[string]string `json:"selector" gorm:"serializer:json"`
```

GORM 需要 `serializer:json` 标签来处理 Go 结构体与 PostgreSQL JSONB 的转换。

**修复的文件**:
- `internal/model/agent.go` - Labels 字段
- `internal/model/configuration.go` - Selector, Platform 字段

---

### 问题 2: TLS 配置错误

**现象**:
```
tls: first record does not look like a TLS handshake
```

**原因**:
`InsecureSkipVerify: true` 只跳过证书验证，仍尝试建立 TLS 连接。
服务器使用 `ws://` (非加密 WebSocket)，不支持 TLS。

**解决方案**:
```go
// ❌ 错误 - 仍使用 TLS
if initialInsecureConnection {
    agent.tlsConfig = &tls.Config{
        InsecureSkipVerify: true,
    }
}

// ✅ 正确 - 完全禁用 TLS
if initialInsecureConnection {
    agent.tlsConfig = nil
}
```

**协议匹配**:
- `ws://` → `tlsConfig = nil`
- `wss://` → `tlsConfig = &tls.Config{...}`

---

### 问题 3: Labels 匹配逻辑错误

**现象**: 空 selector 匹配了所有 Agent

**原因**: 匹配函数未处理空 selector 的情况

**解决方案**:
```go
func (l Labels) Matches(selector map[string]string) bool {
    // 修复：空选择器不匹配任何 Agent
    if len(selector) == 0 {
        return false
    }

    for key, value := range selector {
        if l[key] != value {
            return false
        }
    }
    return true
}
```

---

### 问题 4: Agent 常量命名

**现象**: 编译错误 `undefined: AgentStatusConnected`

**原因**: 常量命名不一致

**解决方案**:
```go
// ❌ 错误
AgentStatusConnected

// ✅ 正确
StatusConnected
```

---

## 📚 术语表

| 术语 | 英文 | 说明 |
|------|------|------|
| **OpAMP** | Open Agent Management Protocol | OpenTelemetry 代理管理协议 |
| **Agent** | Telemetry Agent | 遥测代理，收集观测数据的客户端程序 |
| **Configuration** | Agent Configuration | Agent 的配置信息，包含采集和处理规则 |
| **Selector** | Label Selector | 标签选择器，用于匹配特定的 Agent |
| **Labels** | Labels | Agent 的标签集合，用于分类和匹配 |
| **JSONB** | JSON Binary | PostgreSQL 的二进制 JSON 存储类型 |
| **GORM** | Go ORM | Go 语言的对象关系映射框架 |
| **ADR** | Architecture Decision Record | 架构决策记录 |
| **CI/CD** | Continuous Integration/Delivery | 持续集成/持续交付 |

---

## 🔗 相关链接

- [OpAMP 规范](https://github.com/open-telemetry/opamp-spec)
- [opamp-go 仓库](https://github.com/open-telemetry/opamp-go)
- [GORM 文档](https://gorm.io/docs/)
- [Gin 文档](https://gin-gonic.com/docs/)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)

---

**文档维护**: 当架构、技术栈或关键决策发生变化时，及时更新本文档。
