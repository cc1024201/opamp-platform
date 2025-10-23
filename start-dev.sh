#!/bin/bash

# OpAMP Platform 开发环境启动脚本
#
# 功能:
# 1. 检查依赖
# 2. 启动 Docker 基础设施
# 3. 启动后端服务器
# 4. 启动前端开发服务器

set -e  # 遇到错误立即退出

echo "============================================"
echo "  OpAMP Platform 开发环境启动"
echo "============================================"
echo ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "❌ 错误: Docker 未安装"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# 检查 Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "❌ 错误: Docker Compose 未安装"
    echo "请先安装 Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

# 检查 Go
if ! command -v go &> /dev/null; then
    echo "❌ 错误: Go 未安装"
    echo "请先安装 Go: https://golang.org/doc/install"
    exit 1
fi

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "❌ 错误: Node.js 未安装"
    echo "请先安装 Node.js: https://nodejs.org/"
    exit 1
fi

echo "✅ 依赖检查通过"
echo ""

# 启动 Docker 基础设施
echo "🐳 启动 Docker 基础设施 (PostgreSQL, Redis, MinIO)..."
docker-compose up -d

echo "⏳ 等待数据库就绪..."
sleep 5

# 检查 Docker 容器状态
echo ""
docker-compose ps
echo ""

# 检查后端依赖
echo "📦 检查后端依赖..."
cd backend
if [ ! -d "vendor" ]; then
    echo "安装后端依赖..."
    go mod download
fi

# 启动后端服务器 (后台)
echo ""
echo "🚀 启动后端服务器..."
go run ./cmd/server &
BACKEND_PID=$!
echo "后端服务器 PID: $BACKEND_PID"

# 等待后端启动
echo "⏳ 等待后端服务器启动..."
sleep 3

# 检查后端是否启动成功
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ 后端服务器启动成功: http://localhost:8080"
else
    echo "⚠️  后端服务器可能未完全启动,继续等待..."
    sleep 5
fi

# 创建默认管理员账号
echo ""
echo "👤 创建默认管理员账号..."
go run scripts/create_admin.go 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✅ 管理员账号已就绪"
else
    echo "⚠️  管理员账号创建失败或已存在"
fi

# 进入前端目录
cd ../frontend

# 检查前端依赖
if [ ! -d "node_modules" ]; then
    echo ""
    echo "📦 安装前端依赖 (首次运行)..."
    npm install
fi

# 启动前端开发服务器 (后台)
echo ""
echo "🎨 启动前端开发服务器..."
npm run dev &
FRONTEND_PID=$!
echo "前端服务器 PID: $FRONTEND_PID"

# 等待前端启动
sleep 3

echo ""
echo "============================================"
echo "  🎉 OpAMP Platform 启动完成!"
echo "============================================"
echo ""
echo "📍 服务访问地址:"
echo ""
echo "  前端界面:      http://localhost:3000"
echo "  后端 API:      http://localhost:8080"
echo "  Swagger 文档:  http://localhost:8080/swagger/index.html"
echo "  健康检查:      http://localhost:8080/health"
echo "  Prometheus:    http://localhost:8080/metrics"
echo "  MinIO:         http://localhost:9001"
echo ""
echo "🔑 默认管理员账号:"
echo "  用户名: admin"
echo "  密码:   admin123"
echo ""
echo "⚠️  注意:"
echo "  - 后端服务器 PID: $BACKEND_PID"
echo "  - 前端服务器 PID: $FRONTEND_PID"
echo "  - 停止服务: 按 Ctrl+C 或运行 ./stop-dev.sh"
echo ""
echo "============================================"

# 保存 PID 到文件
cd ..
echo "$BACKEND_PID" > .backend.pid
echo "$FRONTEND_PID" > .frontend.pid

# 等待用户中断
echo ""
echo "按 Ctrl+C 停止所有服务..."
echo ""

# 捕获 Ctrl+C 信号
trap 'echo ""; echo "🛑 停止服务中..."; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; docker-compose down; rm -f .backend.pid .frontend.pid; echo "✅ 所有服务已停止"; exit' INT

# 保持脚本运行
wait
