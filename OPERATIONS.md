# OpAMP Platform 运维手册

**版本**: v1.2.0
**最后更新**: 2025-10-22

本文档提供 OpAMP Platform 的日常运维指南，包括监控、故障排查、备份恢复等。

---

## 📋 目录

1. [日常运维](#日常运维)
2. [监控指标](#监控指标)
3. [告警规则](#告警规则)
4. [备份和恢复](#备份和恢复)
5. [故障排查](#故障排查)
6. [性能优化](#性能优化)
7. [安全最佳实践](#安全最佳实践)

---

## 🔧 日常运维

### 健康检查

OpAMP Platform 提供多个健康检查端点：

```bash
# 详细健康检查
curl http://localhost:8080/health

# Readiness probe（Kubernetes 使用）
curl http://localhost:8080/health/ready

# Liveness probe（Kubernetes 使用）
curl http://localhost:8080/health/live
```

### 查看日志

```bash
# Docker Compose
docker-compose logs -f opamp-server

# Kubernetes
kubectl logs -f deployment/opamp-server -n opamp-platform

# 查看最近 100 行
kubectl logs --tail=100 deployment/opamp-server -n opamp-platform

# 多个 pod 的日志
kubectl logs -l app=opamp-server -n opamp-platform --all-containers=true
```

### 重启服务

```bash
# Docker Compose
docker-compose restart opamp-server

# Kubernetes (滚动重启)
kubectl rollout restart deployment/opamp-server -n opamp-platform

# 查看重启状态
kubectl rollout status deployment/opamp-server -n opamp-platform
```

---

## 📊 监控指标

### Prometheus Metrics

OpAMP Platform 导出以下 Prometheus 指标：

#### HTTP 指标

```
# 请求总数
opamp_platform_http_requests_total{method="GET",path="/health",status="200"}

# 请求延迟
opamp_platform_http_request_duration_seconds{method="GET",path="/api/v1/agents"}

# 请求大小
opamp_platform_http_request_size_bytes{method="POST",path="/api/v1/configurations"}

# 响应大小
opamp_platform_http_response_size_bytes{method="GET",path="/api/v1/agents"}
```

#### Agent 指标

```
# Agent 总数
opamp_platform_agents_total

# 在线 Agent 数
opamp_platform_agents_connected

# 离线 Agent 数
opamp_platform_agents_disconnected

# 连接次数
opamp_platform_agent_connect_total

# 断开次数
opamp_platform_agent_disconnect_total
```

#### 数据库指标

```
# 打开的连接数
opamp_platform_db_connections_open

# 空闲连接数
opamp_platform_db_connections_idle

# 查询总数
opamp_platform_db_queries_total{operation="select"}

# 查询延迟
opamp_platform_db_query_duration_seconds{operation="insert"}
```

### Grafana Dashboard

推荐使用以下 Grafana 面板监控系统：

1. **系统概览**
   - 在线 Agent 数量
   - HTTP 请求 QPS
   - 平均响应时间
   - 错误率

2. **数据库性能**
   - 连接池使用率
   - 查询延迟分布
   - 慢查询监控

3. **资源使用**
   - CPU 使用率
   - 内存使用率
   - 磁盘 I/O

---

## 🚨 告警规则

### Prometheus AlertManager 规则

创建 `prometheus/rules.yml`:

```yaml
groups:
- name: opamp_platform
  interval: 30s
  rules:
  # 服务不可用告警
  - alert: OpAMPServerDown
    expr: up{job="opamp-server"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "OpAMP Server 不可用"
      description: "OpAMP Server {{ $labels.instance }} 已停止响应超过 1 分钟"

  # 高错误率告警
  - alert: HighErrorRate
    expr: |
      sum(rate(opamp_platform_http_requests_total{status=~"5.."}[5m]))
      /
      sum(rate(opamp_platform_http_requests_total[5m]))
      > 0.05
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "HTTP 错误率过高"
      description: "5xx 错误率超过 5% (当前: {{ $value | humanizePercentage }})"

  # 响应时间过长告警
  - alert: HighResponseTime
    expr: |
      histogram_quantile(0.95,
        sum(rate(opamp_platform_http_request_duration_seconds_bucket[5m]))
        by (le, path)
      ) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "API 响应时间过长"
      description: "{{ $labels.path }} 的 P95 响应时间超过 1 秒"

  # Agent 大量离线告警
  - alert: ManyAgentsDisconnected
    expr: opamp_platform_agents_disconnected > 100
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "大量 Agent 离线"
      description: "当前有 {{ $value }} 个 Agent 处于离线状态"

  # 数据库连接池耗尽告警
  - alert: DatabaseConnectionPoolExhausted
    expr: opamp_platform_db_connections_open >= opamp_platform_db_connections_max * 0.9
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "数据库连接池即将耗尽"
      description: "当前使用 {{ $value }} 个连接，接近最大限制"

  # 内存使用过高告警
  - alert: HighMemoryUsage
    expr: |
      (process_resident_memory_bytes / node_memory_MemTotal_bytes) > 0.8
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "内存使用率过高"
      description: "内存使用率超过 80%"
```

### 告警通知配置

在 `alertmanager.yml` 中配置通知渠道：

```yaml
route:
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'default'

receivers:
- name: 'default'
  email_configs:
  - to: 'ops-team@example.com'
    from: 'alertmanager@example.com'
    smarthost: 'smtp.example.com:587'
    auth_username: 'alertmanager@example.com'
    auth_password: 'password'
  webhook_configs:
  - url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
```

---

## 💾 备份和恢复

### 数据库备份

#### 自动备份脚本

创建 `backup.sh`:

```bash
#!/bin/bash

# 配置
BACKUP_DIR="/var/backups/opamp"
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="opamp"
DB_NAME="opamp_platform"
RETENTION_DAYS=7

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份文件名
BACKUP_FILE="$BACKUP_DIR/opamp_$(date +%Y%m%d_%H%M%S).sql.gz"

# 执行备份
PGPASSWORD=$DB_PASSWORD pg_dump \
  -h $DB_HOST \
  -p $DB_PORT \
  -U $DB_USER \
  -d $DB_NAME \
  | gzip > $BACKUP_FILE

# 检查备份是否成功
if [ $? -eq 0 ]; then
    echo "✅ 备份成功: $BACKUP_FILE"

    # 删除旧备份
    find $BACKUP_DIR -name "opamp_*.sql.gz" -mtime +$RETENTION_DAYS -delete
    echo "🧹 已清理 $RETENTION_DAYS 天前的备份"
else
    echo "❌ 备份失败"
    exit 1
fi
```

#### 设置定时备份

```bash
# 添加到 crontab (每天凌晨 2 点备份)
0 2 * * * /path/to/backup.sh >> /var/log/opamp-backup.log 2>&1
```

### 数据恢复

```bash
# 从备份恢复
gunzip < opamp_20251022_020000.sql.gz | psql \
  -h localhost \
  -p 5432 \
  -U opamp \
  -d opamp_platform
```

---

## 🔍 故障排查

### 常见问题

#### 1. 服务无法启动

**症状**: 服务启动失败或立即退出

**排查步骤**:

```bash
# 查看日志
docker-compose logs opamp-server
kubectl logs deployment/opamp-server -n opamp-platform

# 检查配置文件
cat backend/config.yaml

# 检查数据库连接
psql -h localhost -U opamp -d opamp_platform -c "SELECT 1"

# 检查端口占用
netstat -tulpn | grep 8080
```

**常见原因**:
- 数据库连接失败
- 配置文件错误
- 端口被占用
- 权限不足

---

#### 2. Agent 无法连接

**症状**: Agent 报告连接失败

**排查步骤**:

```bash
# 检查 OpAMP 端点是否可访问
curl -v ws://localhost:8080/v1/opamp

# 检查防火墙规则
iptables -L -n | grep 8080

# 检查服务器日志
grep "Agent connect" /var/log/opamp-server.log
```

**常见原因**:
- 网络不通
- 防火墙阻止
- TLS 配置错误
- Secret key 不匹配

---

#### 3. 数据库连接池耗尽

**症状**: API 响应缓慢，数据库连接错误

**排查步骤**:

```bash
# 查看当前连接数
psql -U opamp -d opamp_platform -c "
  SELECT count(*) as connections
  FROM pg_stat_activity
  WHERE datname='opamp_platform';
"

# 查看长时间运行的查询
psql -U opamp -d opamp_platform -c "
  SELECT pid, now() - query_start as duration, query
  FROM pg_stat_activity
  WHERE state = 'active'
  ORDER BY duration DESC;
"
```

**解决方案**:
- 增加连接池大小
- 优化慢查询
- 检查连接泄漏
- 重启服务释放连接

---

#### 4. 内存使用持续增长

**症状**: 内存使用率不断上升

**排查步骤**:

```bash
# 查看内存使用
free -h

# 查看进程内存
ps aux | grep opamp-server

# 使用 pprof 分析内存
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

**解决方案**:
- 检查是否有内存泄漏
- 调整 GOGC 参数
- 增加内存限制
- 定期重启服务

---

## ⚡ 性能优化

### 1. 数据库优化

```sql
-- 创建索引
CREATE INDEX CONCURRENTLY idx_agents_updated_at ON agents(updated_at DESC);
CREATE INDEX CONCURRENTLY idx_configurations_created_at ON configurations(created_at DESC);

-- 分析表统计信息
ANALYZE agents;
ANALYZE configurations;

-- 清理死元组
VACUUM ANALYZE agents;
```

### 2. 缓存策略

```yaml
# config.yaml
redis:
  # 启用缓存
  cache_enabled: true
  # 缓存 TTL
  cache_ttl: 300  # 5 分钟
  # 缓存大小
  max_memory: 256mb
  max_memory_policy: allkeys-lru
```

### 3. 连接池调优

```yaml
database:
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 300  # 5 分钟
```

---

## 🔒 安全最佳实践

### 1. 密钥管理

```bash
# 生成安全的 JWT 密钥
openssl rand -base64 32

# 使用环境变量
export JWT_SECRET_KEY=$(openssl rand -base64 32)

# Kubernetes Secret
kubectl create secret generic opamp-secrets \
  --from-literal=jwt.secret=$(openssl rand -base64 32) \
  -n opamp-platform
```

### 2. 网络安全

```bash
# 限制数据库访问
# /etc/postgresql/16/main/pg_hba.conf
host opamp_platform opamp 10.0.0.0/8 md5

# 配置防火墙
ufw allow 8080/tcp
ufw enable
```

### 3. 定期更新

```bash
# 更新依赖
go get -u ./...
go mod tidy

# 更新 Docker 镜像
docker-compose pull
docker-compose up -d
```

### 4. 审计日志

监控以下关键操作：
- 用户登录/登出
- 配置变更
- Agent 添加/删除
- 权限变更

---

## 📞 紧急联系

- **技术负责人**: [Name] <email@example.com>
- **运维团队**: ops-team@example.com
- **监控告警**: Slack #opamp-alerts

---

**文档维护**: 当系统配置或运维流程发生变化时，及时更新本文档。

**最后更新**: 2025-10-22
