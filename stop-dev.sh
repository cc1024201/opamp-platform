#!/bin/bash

# OpAMP Platform 停止脚本

echo "🛑 停止 OpAMP Platform 开发环境..."
echo ""

# 读取 PID
if [ -f .backend.pid ]; then
    BACKEND_PID=$(cat .backend.pid)
    echo "停止后端服务器 (PID: $BACKEND_PID)..."
    kill $BACKEND_PID 2>/dev/null || echo "后端服务器已停止"
    rm -f .backend.pid
fi

if [ -f .frontend.pid ]; then
    FRONTEND_PID=$(cat .frontend.pid)
    echo "停止前端服务器 (PID: $FRONTEND_PID)..."
    kill $FRONTEND_PID 2>/dev/null || echo "前端服务器已停止"
    rm -f .frontend.pid
fi

# 停止 Docker 容器
echo "停止 Docker 容器..."
docker-compose down

echo ""
echo "✅ 所有服务已停止"
