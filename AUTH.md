# JWT 认证系统使用指南

**最后更新**: 2025-10-22

OpAMP Platform 已集成 JWT (JSON Web Token) 认证系统，保护 API 端点安全。

---

## 🔐 功能特性

- ✅ **JWT 认证** - 基于 token 的无状态认证
- ✅ **用户注册和登录** - 完整的用户管理
- ✅ **密码加密** - 使用 bcrypt 哈希存储
- ✅ **角色管理** - 支持 admin 和 user 角色
- ✅ **API 保护** - 所有业务 API 需要认证
- ✅ **Token 过期** - 可配置的 token 有效期

---

## 🚀 快速开始

### 1. 启动服务

```bash
cd backend

# 启动 PostgreSQL（如果还没启动）
docker-compose up -d postgres

# 启动服务器
./bin/opamp-server
```

### 2. 创建管理员用户

```bash
# 运行初始化脚本
go run scripts/create_admin.go
```

**默认管理员凭证**:
- Username: `admin`
- Password: `admin123`
- Email: `admin@opamp.local`

⚠️ **请在首次登录后立即修改密码！**

### 3. 测试认证系统

```bash
# 运行测试脚本
./test-auth.sh
```

---

## 📚 API 使用指南

### 公开 API（不需要认证）

#### 注册新用户

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "newuser",
    "email": "user@example.com",
    "role": "user",
    "is_active": true
  }
}
```

#### 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@opamp.local",
    "role": "admin",
    "is_active": true
  }
}
```

### 受保护 API（需要认证）

所有业务 API 都需要在请求头中携带 JWT token：

```
Authorization: Bearer <your-token>
```

#### 获取当前用户信息

```bash
curl http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 列出 Agents

```bash
curl http://localhost:8080/api/v1/agents \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 创建配置

```bash
curl -X POST http://localhost:8080/api/v1/configurations \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod-config",
    "display_name": "生产环境配置",
    "content_type": "yaml",
    "raw_config": "receivers:\n  otlp:\n    protocols:\n      grpc:",
    "selector": {
      "env": "prod"
    }
  }'
```

---

## 🔧 配置

在 `config.yaml` 中配置 JWT：

```yaml
jwt:
  # JWT Secret Key (生产环境必须修改为强密钥)
  secret_key: "your-secret-key-change-in-production"
  # Token 有效期
  duration: 24h
```

**重要配置项**:

| 配置项 | 说明 | 默认值 | 生产环境建议 |
|--------|------|--------|-------------|
| `secret_key` | JWT 签名密钥 | 默认密钥 | 使用强随机密钥（至少 32 字节） |
| `duration` | Token 有效期 | 24h | 根据安全需求调整 |

### 生成安全密钥

```bash
# 方法 1: 使用 openssl
openssl rand -base64 32

# 方法 2: 使用 Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

---

## 🛡️ 安全最佳实践

### 1. 密钥管理

- ✅ 生产环境使用环境变量存储密钥
- ✅ 定期轮换 JWT 密钥
- ❌ 不要将密钥提交到版本控制

```bash
# 使用环境变量
export JWT_SECRET_KEY="your-production-secret-key"
```

### 2. Token 管理

- ✅ Token 应该存储在安全的地方（如 HttpOnly Cookie）
- ✅ 实现 token 刷新机制
- ✅ 登出时清除 token

### 3. HTTPS

- ✅ 生产环境必须使用 HTTPS
- ❌ 不要在 HTTP 上传输 token

### 4. 密码策略

- ✅ 强制使用强密码（最小长度 6 字符）
- ✅ 密码使用 bcrypt 哈希存储
- ✅ 实现密码重置功能

---

## 🔍 错误处理

### 常见错误响应

#### 401 Unauthorized

```json
{
  "error": "authorization header is not provided"
}
```

**原因**: 请求未携带 Authorization 头

**解决**: 在请求头中添加 `Authorization: Bearer <token>`

---

#### 401 Unauthorized

```json
{
  "error": "invalid token"
}
```

**原因**: Token 无效或已过期

**解决**: 重新登录获取新 token

---

#### 403 Forbidden

```json
{
  "error": "insufficient permissions"
}
```

**原因**: 用户角色权限不足

**解决**: 使用具有相应权限的账户

---

## 🧪 前端集成示例

### JavaScript/TypeScript

```typescript
// 登录
async function login(username: string, password: string) {
  const response = await fetch('http://localhost:8080/api/v1/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password }),
  });

  const data = await response.json();

  // 保存 token
  localStorage.setItem('token', data.token);

  return data;
}

// 调用受保护的 API
async function fetchAgents() {
  const token = localStorage.getItem('token');

  const response = await fetch('http://localhost:8080/api/v1/agents', {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  return await response.json();
}
```

### React Hook 示例

```typescript
import { useState, useEffect } from 'react';

function useAuth() {
  const [token, setToken] = useState<string | null>(
    localStorage.getItem('token')
  );

  const login = async (username: string, password: string) => {
    const response = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });

    const data = await response.json();
    setToken(data.token);
    localStorage.setItem('token', data.token);
  };

  const logout = () => {
    setToken(null);
    localStorage.removeItem('token');
  };

  return { token, login, logout };
}
```

---

## 📊 数据库 Schema

### users 表

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,  -- bcrypt 哈希
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🚧 未来改进

- [ ] 实现 Refresh Token 机制
- [ ] 添加密码重置功能
- [ ] 实现两步验证（2FA）
- [ ] 添加登录历史记录
- [ ] 实现账户锁定机制
- [ ] 添加更细粒度的 RBAC 权限

---

## 🔗 相关文档

- [README.md](README.md) - 项目主页
- [DEVELOPMENT.md](DEVELOPMENT.md) - 开发指南
- [TESTING.md](TESTING.md) - 测试指南

---

**文档维护**: 当认证系统有更新时，及时更新本文档。
