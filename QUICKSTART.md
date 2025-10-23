# OpAMP Platform 快速启动指南

**版本**: v1.1.0 (Phase 2.5)
**最后更新**: 2025-10-22

本指南帮助你快速启动并测试 OpAMP Platform 的新功能。

---

## 🚀 快速启动（5 分钟）

### 1. 启动依赖服务

```bash
# 在项目根目录
cd opamp-platform

# 启动 PostgreSQL, Redis, MinIO
docker-compose up -d

# 验证服务状态
docker-compose ps
```

**预期输出**:
```
NAME                COMMAND                  SERVICE      STATUS
opamp-postgres      "docker-entrypoint..."   postgres     Up 30 seconds
```

### 2. 编译并启动服务器

```bash
cd backend

# 编译
go build -o bin/opamp-server ./cmd/server

# 启动服务器
./bin/opamp-server
```

**预期输出**:
```
INFO  PostgreSQL store initialized
INFO  OpAMP server starting on /v1/opamp
INFO  Server starting  {"port": 8080}
```

### 3. 创建管理员用户

**新终端**:
```bash
cd backend

# 运行管理员创建脚本
go run scripts/create_admin.go
```

**预期输出**:
```
✅ Admin user created successfully!

=== Default Admin Credentials ===
Username: admin
Password: admin123
Email: admin@opamp.local

⚠️  Please change the password after first login!
```

### 4. 测试新功能

```bash
# 在项目根目录
./test-auth.sh
```

---

## 🧪 功能测试

### 测试 1: JWT 认证

#### 注册新用户
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "test123"
  }'
```

**预期响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 2,
    "username": "testuser",
    "email": "test@example.com",
    "role": "user",
    "is_active": true
  }
}
```

#### 登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

保存返回的 `token`，后续请求需要使用。

#### 访问受保护的 API
```bash
# 替换 YOUR_TOKEN 为上面获取的 token
export TOKEN="YOUR_TOKEN"

# 获取当前用户信息
curl http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer $TOKEN"

# 列出 Agents
curl http://localhost:8080/api/v1/agents \
  -H "Authorization: Bearer $TOKEN"
```

---

### 测试 2: 健康检查

#### 详细健康检查
```bash
curl http://localhost:8080/health | jq
```

**预期响应**:
```json
{
  "status": "healthy",
  "timestamp": 1729608600,
  "version": "1.0.0",
  "components": {
    "database": {
      "status": "healthy",
      "message": "database connection successful",
      "latency": "2.5ms"
    },
    "opamp": {
      "status": "healthy",
      "message": "OpAMP server is running"
    }
  }
}
```

#### Kubernetes 探针
```bash
# Readiness probe
curl http://localhost:8080/health/ready

# Liveness probe
curl http://localhost:8080/health/live
```

---

### 测试 3: Prometheus Metrics

```bash
curl http://localhost:8080/metrics
```

**部分指标示例**:
```
# HELP opamp_platform_http_requests_total Total number of HTTP requests
# TYPE opamp_platform_http_requests_total counter
opamp_platform_http_requests_total{method="GET",path="/health",status="200"} 5

# HELP opamp_platform_agents_total Total number of agents
# TYPE opamp_platform_agents_total gauge
opamp_platform_agents_total 0

# HELP opamp_platform_http_request_duration_seconds HTTP request duration in seconds
# TYPE opamp_platform_http_request_duration_seconds histogram
opamp_platform_http_request_duration_seconds_bucket{method="GET",path="/health",le="0.005"} 5
```

---

## 📊 可视化监控（可选）

### 使用 Prometheus 和 Grafana

#### 1. 启动 Prometheus

创建 `prometheus.yml`:
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'opamp-platform'
    static_configs:
      - targets: ['192.168.31.46:8080']  # 使用局域网 IP
```

启动 Prometheus:
```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

访问: http://localhost:9090

#### 2. 启动 Grafana

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  grafana/grafana
```

访问: http://localhost:3000
- 默认用户名/密码: admin/admin

**配置数据源**:
1. 添加 Prometheus 数据源
2. URL: http://192.168.31.46:9090

**导入仪表盘**:
- 搜索 "Gin" 或 "HTTP" 相关的 Grafana 仪表盘

---

## 🎯 常见操作

### 查看日志
```bash
# 服务器日志
./bin/opamp-server

# Docker 服务日志
docker-compose logs -f postgres
```

### 停止服务
```bash
# 停止服务器：Ctrl+C

# 停止 Docker 服务
docker-compose down
```

### 重置数据库
```bash
# 停止服务器
docker-compose down -v  # 删除卷

# 重新启动
docker-compose up -d
```

---

## 🔧 配置

### 修改 JWT 密钥

编辑 `backend/config.yaml`:
```yaml
jwt:
  secret_key: "your-strong-secret-key-here"
  duration: 24h  # Token 有效期
```

### 修改服务端口

编辑 `backend/config.yaml`:
```yaml
server:
  port: 8080
  mode: debug  # 或 release
```

---

## 🐛 故障排查

### 问题 1: 数据库连接失败

**错误信息**: `Failed to connect to database`

**解决方案**:
```bash
# 检查 PostgreSQL 是否运行
docker-compose ps postgres

# 查看日志
docker-compose logs postgres

# 重启数据库
docker-compose restart postgres
```

### 问题 2: 认证失败

**错误信息**: `authorization header is not provided`

**解决方案**:
```bash
# 确保请求头包含 Authorization
curl http://localhost:8080/api/v1/agents \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 问题 3: Token 过期

**错误信息**: `token has expired`

**解决方案**:
```bash
# 重新登录获取新 token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

---

## 📚 下一步

- 📖 阅读 [AUTH.md](AUTH.md) 了解认证系统详情
- 🧪 查看 [TESTING.md](TESTING.md) 学习如何运行测试
- 🚀 查看 [DEVELOPMENT.md](DEVELOPMENT.md) 了解开发指南
- 📊 查看 [PROGRESS.md](PROGRESS.md) 了解最新进展

---

## 🆘 获取帮助

- 问题反馈: GitHub Issues
- 文档: [README.md](README.md)

---

**祝你使用愉快！** 🎉
