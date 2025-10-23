# OpAMP Platform - 前端

OpAMP Agent 管理平台的 Web 前端界面。

## 技术栈

- **React 19** - UI 框架
- **TypeScript 5** - 类型安全
- **Vite 7** - 构建工具
- **Material-UI (MUI) v7** - UI 组件库
- **Zustand** - 状态管理
- **React Router v7** - 路由管理
- **Axios** - HTTP 客户端
- **Monaco Editor** - 代码编辑器 (YAML/JSON)
- **Recharts** - 图表库
- **date-fns** - 日期处理

## 功能特性

### 已实现功能 ✅

- ✅ 用户认证 (登录/注册)
- ✅ JWT Token 管理
- ✅ Agent 列表和详情页面
- ✅ Configuration 管理 (CRUD)
- ✅ YAML/JSON 编辑器
- ✅ 实时状态 Dashboard
- ✅ 响应式布局
- ✅ 路由守卫
- ✅ API 代理配置

## 快速开始

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

访问: http://localhost:3000

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## 项目结构

```
frontend/
├── src/
│   ├── components/         # React 组件
│   │   ├── layout/         # 布局组件
│   │   │   └── MainLayout.tsx
│   │   └── auth/           # 认证组件
│   │       └── ProtectedRoute.tsx
│   ├── pages/              # 页面组件
│   │   ├── auth/           # 认证页面
│   │   │   ├── LoginPage.tsx
│   │   │   └── RegisterPage.tsx
│   │   ├── dashboard/      # 仪表盘
│   │   │   └── DashboardPage.tsx
│   │   ├── agents/         # Agent 管理
│   │   │   ├── AgentListPage.tsx
│   │   │   └── AgentDetailPage.tsx
│   │   └── configurations/ # 配置管理
│   │       └── ConfigurationListPage.tsx
│   ├── services/           # API 服务
│   │   ├── api.ts          # Axios 实例
│   │   ├── auth.service.ts
│   │   ├── agent.service.ts
│   │   └── configuration.service.ts
│   ├── stores/             # Zustand 状态管理
│   │   ├── authStore.ts
│   │   ├── agentStore.ts
│   │   └── configurationStore.ts
│   ├── types/              # TypeScript 类型定义
│   │   └── api.ts
│   ├── App.tsx             # 主应用组件
│   └── main.tsx            # 入口文件
├── public/                 # 静态资源
├── vite.config.ts          # Vite 配置
├── tsconfig.json           # TypeScript 配置
└── package.json            # 项目依赖
```

## 配置说明

### API 代理

前端开发服务器配置了 API 代理,自动将请求转发到后端服务器:

- `/api/*` → `http://localhost:8080`
- `/v1/*` → `http://localhost:8080` (支持 WebSocket)

配置文件: `vite.config.ts`

### 路径别名

配置了 `@` 别名指向 `src` 目录:

```typescript
import { useAuthStore } from '@/stores/authStore';
import { Agent } from '@/types/api';
```

## 使用指南

### 1. 登录系统

首次使用需要先登录或注册账号。

默认管理员账号:
- 用户名: `admin`
- 密码: `admin123`

### 2. 查看 Dashboard

登录后可以看到:
- Agent 总数统计
- 在线/离线 Agent 数量
- 配置总数
- 最近连接的 Agents
- 最近更新的配置

### 3. 管理 Agents

- 查看所有 Agent 列表
- 点击查看 Agent 详情
- 删除 Agent
- 支持分页和刷新

### 4. 管理 Configurations

- 查看配置列表
- 创建新配置 (支持 YAML/JSON)
- 编辑现有配置
- 删除配置
- 配置标签选择器

## 开发说明

### 添加新页面

1. 在 `src/pages/` 创建新页面组件
2. 在 `src/App.tsx` 添加路由
3. 在 `src/components/layout/MainLayout.tsx` 添加菜单项

### 添加新 API

1. 在 `src/types/api.ts` 定义类型
2. 在 `src/services/` 创建服务
3. 在 `src/stores/` 创建状态管理 (如需要)

### 状态管理

使用 Zustand 进行状态管理:

```typescript
// 使用 store
const { user, login, logout } = useAuthStore();

// 调用 action
await login(username, password);
```

## 注意事项

- 确保后端服务器运行在 `http://localhost:8080`
- Token 存储在 `localStorage` 中
- Token 过期会自动跳转到登录页
- 所有业务 API 需要 JWT 认证

## 未来计划

- [ ] 添加实时 WebSocket 连接 (Agent 状态实时更新)
- [ ] 添加图表可视化 (Agent 状态分布、连接趋势)
- [ ] 添加单元测试
- [ ] 添加 E2E 测试
- [ ] 性能优化 (代码分割、懒加载)
- [ ] PWA 支持
- [ ] 暗黑模式
- [ ] 国际化 (i18n)

## License

MIT
