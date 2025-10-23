# OpAMP Platform 部署指南

**版本**: v1.2.0
**最后更新**: 2025-10-22

本文档提供 OpAMP Platform 的完整部署指南，包括开发环境、生产环境和 Kubernetes 部署。

---

## 📋 目录

1. [系统要求](#系统要求)
2. [开发环境部署](#开发环境部署)
3. [生产环境部署](#生产环境部署)
4. [Kubernetes 部署](#kubernetes-部署)
5. [配置管理](#配置管理)
6. [监控和告警](#监控和告警)
7. [备份和恢复](#备份和恢复)
8. [故障排查](#故障排查)

---

## 🔧 系统要求

### 最低配置
- **CPU**: 2 核
- **内存**: 4GB
- **磁盘**: 20GB
- **操作系统**: Linux/macOS/Windows

### 推荐配置（生产环境）
- **CPU**: 4 核+
- **内存**: 8GB+
- **磁盘**: 100GB+ SSD
- **操作系统**: Linux (Ubuntu 20.04+ / CentOS 8+)

### 依赖软件
- **Go**: 1.24+
- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **PostgreSQL**: 16+
- **Redis**: 7+
- **MinIO**: latest

---

## 🚀 开发环境部署

### 方式 1: 使用 Makefile（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/cc1024201/opamp-platform.git
cd opamp-platform/backend

# 2. 一键初始化环境
make setup

# 3. 启动服务器
make run
```

**make setup 会自动完成**:
- ✅ 安装 Go 依赖
- ✅ 启动 Docker 服务（PostgreSQL、Redis、MinIO）
- ✅ 运行数据库迁移
- ✅ 创建管理员用户

### 方式 2: 手动部署

```bash
# 1. 启动依赖服务
cd opamp-platform
docker-compose up -d

# 2. 编译项目
cd backend
go build -o bin/opamp-server ./cmd/server

# 3. 运行数据库迁移
make migrate-up

# 4. 创建管理员
make create-admin

# 5. 启动服务器
./bin/opamp-server
```

### 验证部署

访问以下地址验证部署是否成功：

- **API 文档**: http://localhost:8080/swagger/index.html
- **健康检查**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics

---

## 🏭 生产环境部署

### 1. 使用 Docker Compose（单机部署）

#### 创建生产配置

创建 `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: opamp-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: opamp
      POSTGRES_PASSWORD: ${DB_PASSWORD}  # 使用环境变量
      POSTGRES_DB: opamp_platform
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U opamp"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: opamp-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  minio:
    image: minio/minio:latest
    container_name: opamp-minio
    restart: unless-stopped
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ACCESS_KEY}
      MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY}
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

  opamp-server:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: opamp-server
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=opamp
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=opamp_platform
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - MINIO_ENDPOINT=minio:9000
      - MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY}
      - MINIO_SECRET_KEY=${MINIO_SECRET_KEY}
      - JWT_SECRET_KEY=${JWT_SECRET_KEY}
      - SERVER_MODE=release
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:
  redis_data:
  minio_data:
```

#### 创建 Dockerfile

创建 `backend/Dockerfile`:

```dockerfile
# 构建阶段
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git make

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o opamp-server ./cmd/server

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/opamp-server .
COPY --from=builder /app/config.yaml .

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./opamp-server"]
