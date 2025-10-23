# 🔄 Phase 4.3 进行中 - Agent 状态管理增强

**开始日期**: 2025-10-23
**完成日期**: 2025-10-23
**版本**: v2.3.0-alpha
**状态**: ✅ 核心功能已完成

---

## 📊 项目概述

实现 Agent 状态管理增强功能,包括心跳监控、连接状态持久化、离线检测等核心功能,提升系统对 Agent 的可观测性和可靠性。

---

## 🎯 目标

### 核心功能
1. **Agent 心跳监控** - 实时监控 Agent 健康状态
2. **连接状态持久化** - 持久化 Agent 连接历史
3. **离线 Agent 处理** - 自动检测和处理离线 Agent
4. **Agent 元数据完整性** - 确保 Agent 信息的准确性

### 预期成果
- Agent 状态实时可见
- 离线 Agent 自动检测 (超时时间可配置)
- Agent 连接历史可追溯
- 完善的状态转换日志

---

## 📋 任务清单

### 1. 数据库设计 ✅
- [x] 设计 `agent_connection_history` 表
  - 记录 Agent 连接/断开历史
  - 记录连接时长、断开原因
- [x] 为 `agents` 表添加状态字段
  - `status`: online/offline/error
  - `last_seen_at`: 最后心跳时间
  - `last_connected_at`: 最后连接时间
  - `last_disconnected_at`: 最后断开时间
- [x] 创建 migration 文件
  - `migrations/000004_add_agent_status_tracking.up.sql`
  - `migrations/000004_add_agent_status_tracking.down.sql`

### 2. 后端实现

#### 2.1 数据模型 ✅
- [x] 更新 `internal/model/agent.go`
  - 将 AgentStatus 改为 string 类型
  - 添加 StatusOnline, StatusOffline, StatusError 常量
  - 添加 LastSeenAt, LastConnectedAt, LastDisconnectedAt 字段
  - 添加 DisconnectReason 字段
- [x] 创建 `internal/model/agent_connection_history.go`
  - AgentConnectionHistory 模型
  - CalculateDuration 方法

#### 2.2 数据访问层 ✅
- [x] 创建 `internal/store/postgres/agent_connection_history.go`
  - CreateConnectionHistory - 创建连接历史
  - UpdateConnectionHistory - 更新连接历史
  - GetActiveConnectionHistory - 获取活跃连接
  - ListConnectionHistoryByAgent - 列出 Agent 连接历史
  - UpdateAgentStatus - 更新 Agent 状态
  - UpdateAgentLastSeen - 更新心跳时间
  - SetAgentDisconnectReason - 设置断开原因
  - ListOnlineAgents - 列出在线 Agent
  - ListOfflineAgents - 列出离线 Agent
  - ListStaleAgents - 列出心跳超时的 Agent

#### 2.3 心跳监控 ✅
- [x] 创建 `internal/opamp/heartbeat_monitor.go`
  - HeartbeatMonitor 结构体
  - Start/Stop 方法
  - checkHeartbeats - 定期检查心跳
  - handleHeartbeatTimeout - 处理超时
  - 默认配置: 30秒检查间隔, 60秒超时

#### 2.4 OpAMP 集成 ✅
- [x] 更新 `internal/opamp/server.go`
  - 扩展 AgentStore 接口
  - 集成 HeartbeatMonitor
  - Start 时启动心跳监控
  - Stop 时停止心跳监控
- [x] 更新 `internal/opamp/callbacks.go`
  - onMessage: 更新 Agent 心跳时间
  - onConnectionClose: 更新状态为离线,记录连接历史
  - updateAgentState: Agent 上线时创建连接历史

#### 2.5 状态查询 API
- [ ] 添加 API 端点 (待实现)
  - `GET /api/v1/agents/:id/status` - 获取 Agent 状态详情
  - `GET /api/v1/agents/:id/connection-history` - 查看连接历史
  - `GET /api/v1/agents/offline` - 列出所有离线 Agent
  - `POST /api/v1/agents/:id/reconnect` - 触发 Agent 重连

### 3. 监控和指标
- [ ] 添加 Prometheus metrics (待实现)
  - `opamp_agents_online_total` - 在线 Agent 数量
  - `opamp_agents_offline_total` - 离线 Agent 数量
  - `opamp_agent_connection_duration_seconds` - 连接时长分布
  - `opamp_agent_heartbeat_missed_total` - 心跳超时次数

### 4. 测试 ✅
- [x] 单元测试
  - ✅ 更新 mock 实现以支持新接口
  - ✅ 更新状态常量 (StatusConnected → StatusOnline)
  - ✅ 所有 opamp 模块测试通过
- [ ] 集成测试 (待实现)
  - Agent 连接/断开完整流程
  - 心跳超时自动检测
  - 状态查询 API 测试

---

## 🏗️ 技术设计

### 状态机设计

```
┌─────────────────────────────────────────────┐
│           Agent 状态转换图                   │
└─────────────────────────────────────────────┘

    [初始]
       │
       ↓ (Agent 连接)
   ┌────────┐
   │ Online │ ←──────────────┐
   └────────┘                │
       │                     │
       │ (心跳超时)          │ (Agent 重连)
       ↓                     │
   ┌────────┐                │
   │Offline │ ───────────────┘
   └────────┘
       │
       │ (持续离线)
       ↓
   ┌────────┐
   │ Error  │ (可能需要人工介入)
   └────────┘
```

### 数据库表设计

#### agents 表新增字段
```sql
ALTER TABLE agents ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'offline';
ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMP;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_connected_at TIMESTAMP;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_disconnected_at TIMESTAMP;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS disconnect_reason TEXT;
```

#### agent_connection_history 表
```sql
CREATE TABLE agent_connection_history (
    id BIGSERIAL PRIMARY KEY,
    agent_id VARCHAR(255) NOT NULL,
    connected_at TIMESTAMP NOT NULL,
    disconnected_at TIMESTAMP,
    duration_seconds INTEGER,
    disconnect_reason TEXT,
    remote_addr VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_agent
        FOREIGN KEY (agent_id)
        REFERENCES agents(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_agent_connection_history_agent_id ON agent_connection_history(agent_id);
CREATE INDEX idx_agent_connection_history_connected_at ON agent_connection_history(connected_at);
```

### 心跳监控实现

```go
// HeartbeatMonitor 心跳监控器
type HeartbeatMonitor struct {
    store          AgentStore
    checkInterval  time.Duration
    timeout        time.Duration
    stopCh         chan struct{}
}

func (m *HeartbeatMonitor) Start() {
    ticker := time.NewTicker(m.checkInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            m.checkHeartbeats()
        case <-m.stopCh:
            return
        }
    }
}

func (m *HeartbeatMonitor) checkHeartbeats() {
    // 查询所有在线 Agent
    agents := m.store.ListOnlineAgents()

    now := time.Now()
    for _, agent := range agents {
        if now.Sub(agent.LastSeenAt) > m.timeout {
            // 心跳超时,标记为离线
            m.handleHeartbeatTimeout(agent)
        }
    }
}
```

---

## 🔧 实现步骤

### 第一步: 数据库迁移 (预计 30 分钟)
1. 创建 migration 文件
2. 测试 up/down 迁移
3. 更新数据模型

### 第二步: 状态管理器 (预计 1-2 小时)
1. 实现状态转换逻辑
2. 添加状态持久化
3. 编写单元测试

### 第三步: 心跳监控 (预计 1-2 小时)
1. 实现心跳检测逻辑
2. 集成到 OpAMP 回调
3. 添加后台监控任务
4. 编写测试

### 第四步: 连接历史 (预计 1 小时)
1. 实现数据访问层
2. 集成到连接/断开流程
3. 编写测试

### 第五步: API 和监控 (预计 1 小时)
1. 添加状态查询 API
2. 添加 Prometheus metrics
3. 编写 API 测试

### 第六步: 集成测试 (预计 1 小时)
1. 编写端到端测试
2. 测试各种场景
3. 性能测试

**预计总时间**: 6-8 小时

---

## 📈 成功指标

### 功能指标
- [x] Agent 状态准确反映实时连接状态
- [ ] 离线 Agent 在 60 秒内被检测到
- [ ] 连接历史完整记录
- [ ] API 响应时间 < 100ms

### 测试指标
- [ ] 单元测试覆盖率 > 80%
- [ ] 所有集成测试通过
- [ ] 心跳检测准确率 100%

### 性能指标
- [ ] 支持 1000+ Agent 并发监控
- [ ] 心跳检测延迟 < 5 秒
- [ ] 数据库查询性能优化

---

## 📝 技术债务和风险

### 技术债务
- `internal/packagemgr` 和 `internal/storage` 模块缺少测试
- `cmd/server` 模块测试间歇性失败

### 潜在风险
1. **性能风险**: 大量 Agent 时心跳检测可能影响性能
   - 缓解措施: 使用批量查询,添加索引
2. **时钟同步**: 分布式环境时钟不同步
   - 缓解措施: 使用数据库时间戳
3. **数据库锁竞争**: 频繁状态更新
   - 缓解措施: 使用乐观锁或批量更新

---

## 🔗 相关文档

- [ROADMAP.md](ROADMAP.md) - 项目路线图
- [PHASE4_COMPLETED.md](PHASE4_COMPLETED.md) - Phase 4.1 完成报告
- [PHASE4.2_COMPLETED.md](PHASE4.2_COMPLETED.md) - Phase 4.2 完成报告
- [OpAMP 协议规范](https://github.com/open-telemetry/opamp-spec)

---

## 📅 时间线

| 日期 | 任务 | 状态 |
|------|------|------|
| 2025-10-23 | 创建 Phase 4.3 文档 | ✅ 完成 |
| 2025-10-23 | 设计数据库表结构 | ✅ 完成 |
| 2025-10-23 | 实现数据模型 | ✅ 完成 |
| 2025-10-23 | 实现数据访问层 | ✅ 完成 |
| 2025-10-23 | 实现心跳监控 | ✅ 完成 |
| 2025-10-23 | 集成 OpAMP 回调 | ✅ 完成 |
| 2025-10-23 | 更新测试 | ✅ 完成 |
| TBD | 添加状态查询 API | 📋 待开始 |
| TBD | 添加 Prometheus metrics | 📋 待开始 |
| TBD | 完成集成测试 | 📋 待开始 |

---

**当前状态**: ✅ 核心功能已完成
**下一步**: 添加状态查询 API 和监控指标

**负责人**: 开发团队
**优先级**: ⭐⭐⭐ 高优先级
