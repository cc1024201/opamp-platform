#!/bin/bash

# OpAMP Platform 认证测试脚本

BASE_URL="http://localhost:8080/api/v1"

echo "=== OpAMP Platform 认证测试 ==="
echo ""

# 测试 1: 注册新用户
echo "1. 测试注册新用户..."
REGISTER_RESPONSE=$(curl -s -X POST ${BASE_URL}/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "test123"
  }')
echo "注册响应: $REGISTER_RESPONSE"
echo ""

# 提取 token
TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"token":"[^"]*' | sed 's/"token":"//')
echo "获取到的 Token: $TOKEN"
echo ""

# 测试 2: 登录
echo "2. 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST ${BASE_URL}/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')
echo "登录响应: $LOGIN_RESPONSE"
echo ""

# 提取 admin token
ADMIN_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | sed 's/"token":"//')
echo "Admin Token: $ADMIN_TOKEN"
echo ""

# 测试 3: 不带 token 访问受保护的 API（应该失败）
echo "3. 测试不带 token 访问受保护的 API..."
curl -s ${BASE_URL}/agents
echo ""
echo ""

# 测试 4: 带 token 访问受保护的 API（应该成功）
echo "4. 测试带 token 访问受保护的 API..."
if [ -n "$ADMIN_TOKEN" ]; then
  curl -s ${BASE_URL}/agents \
    -H "Authorization: Bearer $ADMIN_TOKEN"
  echo ""
else
  echo "无法获取 token，跳过测试"
fi
echo ""

# 测试 5: 获取当前用户信息
echo "5. 测试获取当前用户信息..."
if [ -n "$ADMIN_TOKEN" ]; then
  curl -s ${BASE_URL}/me \
    -H "Authorization: Bearer $ADMIN_TOKEN"
  echo ""
else
  echo "无法获取 token，跳过测试"
fi
echo ""

echo "=== 测试完成 ==="
