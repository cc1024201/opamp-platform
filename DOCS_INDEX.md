# OpAMP Platform 文档索引

**版本**: v1.2.0 (Phase 2.5)
**最后更新**: 2025-10-23

欢迎来到 OpAMP Platform 文档中心！本文档将帮助您快速找到所需的信息。

---

## 📚 文档分类

### 🚀 快速入门
如果您是第一次接触本项目,从这里开始:

| 文档 | 说明 | 适合对象 |
|------|------|---------|
| **[README.md](README.md)** | 项目主页、概览、快速开始 | 所有人 |
| **[QUICKSTART.md](QUICKSTART.md)** | 5 分钟快速启动指南 | 新用户 |

### 👨‍💻 开发文档
深入了解项目架构和开发细节:

| 文档 | 说明 | 内容 |
|------|------|------|
| **[DEVELOPMENT.md](DEVELOPMENT.md)** | 开发指南 | 架构设计、技术决策、常见问题 |
| **[AUTH.md](AUTH.md)** | 认证系统使用指南 | JWT、用户管理、API 保护 |
| **[TESTING.md](TESTING.md)** | 测试指南 | 测试策略、覆盖率、如何测试 |
| **[backend/TEST_SUMMARY.md](backend/TEST_SUMMARY.md)** | Handler 测试总结 | Phase 2.5 新增的 API 测试 |

### 🚀 部署运维
将项目部署到生产环境:

| 文档 | 说明 | 内容 |
|------|------|------|
| **[DEPLOYMENT.md](DEPLOYMENT.md)** | 部署指南 | Docker、Kubernetes 部署方案 |
| **[OPERATIONS.md](OPERATIONS.md)** | 运维手册 | 监控、告警、备份、故障排查 |

### 📊 项目状态
了解项目的开发进度和完成情况:

| 文档 | 说明 | 状态 |
|------|------|------|
| **[PROJECT_STATUS.md](PROJECT_STATUS.md)** | 项目状态报告 | ✨ 最新 - 75% 完成 (6/8 任务) |

### 📖 API 文档
交互式 API 文档:

| 文档 | 访问方式 | 说明 |
|------|---------|------|
| **Swagger UI** | http://localhost:8080/swagger/index.html | 在线 API 文档(服务运行时) |
| **[docs/swagger.json](backend/docs/swagger.json)** | 文件 | OpenAPI 3.0 JSON 格式 |
| **[docs/swagger.yaml](backend/docs/swagger.yaml)** | 文件 | OpenAPI 3.0 YAML 格式 |

---

## 🎯 按角色导航

### 🆕 新开发者

**学习路径**:
1. 📖 阅读 [README.md](README.md) 了解项目概况
2. 🚀 跟随 [QUICKSTART.md](QUICKSTART.md) 启动项目
3. 🔧 学习 [DEVELOPMENT.md](DEVELOPMENT.md) 了解架构
4. 🧪 参考 [TESTING.md](TESTING.md) 学习测试

**常见问题**:
- 如何启动项目? → [QUICKSTART.md](QUICKSTART.md)
- 项目架构是什么? → [DEVELOPMENT.md](DEVELOPMENT.md)
- 如何运行测试? → [TESTING.md](TESTING.md)

### 👨‍💻 贡献者

**工作流程**:
1. 🔐 了解 [AUTH.md](AUTH.md) - 认证系统实现
2. 📊 查看 [backend/TEST_SUMMARY.md](backend/TEST_SUMMARY.md) - 当前测试状态
3. 🛠️ 参考 [DEVELOPMENT.md](DEVELOPMENT.md) - 常见问题和解决方案
4. 🧪 编写测试 [TESTING.md](TESTING.md) - 测试最佳实践

**开发资源**:
- API 文档: http://localhost:8080/swagger/index.html
- 健康检查: http://localhost:8080/health
- Metrics: http://localhost:8080/metrics

### 🚀 运维人员

**运维资源**:
1. 📦 阅读 [DEPLOYMENT.md](DEPLOYMENT.md) - 部署方案
2. 🔧 参考 [OPERATIONS.md](OPERATIONS.md) - 运维手册
3. 💚 监控 Health Checks - http://localhost:8080/health
4. 📈 查看 Metrics - http://localhost:8080/metrics

**常见任务**:
- 部署到 Docker? → [DEPLOYMENT.md](DEPLOYMENT.md#docker-部署)
- 部署到 Kubernetes? → [DEPLOYMENT.md](DEPLOYMENT.md#kubernetes-部署)
- 配置监控? → [OPERATIONS.md](OPERATIONS.md#监控配置)
- 故障排查? → [OPERATIONS.md](OPERATIONS.md#故障排查)

### 📊 项目经理

**状态报告**:
1. 📈 [PROJECT_STATUS.md](PROJECT_STATUS.md) - 最新项目状态
2. 📖 [README.md](README.md#项目统计) - 项目统计数据

**关键指标**:
- 完成度: 75% (6/8 任务)
- 生产就绪度: 90%
- 测试覆盖率: 38.1%
- 代码行数: ~7,300 行

---

## 📑 按主题导航

### 认证与安全
- [AUTH.md](AUTH.md) - JWT 认证系统完整指南
- [README.md](README.md#认证-api) - 认证 API 快速参考
- [OPERATIONS.md](OPERATIONS.md#安全配置) - 生产环境安全配置

### 测试与质量
- [TESTING.md](TESTING.md) - 完整测试指南
- [backend/TEST_SUMMARY.md](backend/TEST_SUMMARY.md) - Handler 层测试报告
- [README.md](README.md#测试) - 快速测试命令

### 部署与运维
- [DEPLOYMENT.md](DEPLOYMENT.md) - Docker & Kubernetes 部署
- [OPERATIONS.md](OPERATIONS.md) - 监控、告警、备份
- [QUICKSTART.md](QUICKSTART.md) - 快速启动(开发环境)

### API 文档
- **Swagger UI**: http://localhost:8080/swagger/index.html (推荐)
- [README.md](README.md#api-文档) - API 使用示例
- [AUTH.md](AUTH.md#api-示例) - 认证 API 示例

---

## 🔍 快速查找

### 我想知道...

#### "如何快速启动项目?"
→ [QUICKSTART.md](QUICKSTART.md) 或 [README.md#快速开始](README.md#快速开始)

#### "项目使用了哪些技术?"
→ [README.md#技术栈](README.md#技术栈)

#### "如何使用 JWT 认证?"
→ [AUTH.md](AUTH.md)

#### "如何运行测试?"
→ [TESTING.md#如何运行测试](TESTING.md#如何运行测试)

#### "当前测试覆盖率是多少?"
→ [README.md#测试统计](README.md#测试统计) 或 [PROJECT_STATUS.md](PROJECT_STATUS.md#测试覆盖率详情)

#### "如何部署到 Kubernetes?"
→ [DEPLOYMENT.md#kubernetes-部署](DEPLOYMENT.md#kubernetes-部署)

#### "如何配置监控?"
→ [OPERATIONS.md#监控配置](OPERATIONS.md#监控配置)

#### "项目开发进度如何?"
→ [PROJECT_STATUS.md](PROJECT_STATUS.md)

#### "有哪些 API 端点?"
→ http://localhost:8080/swagger/index.html (运行时) 或 [README.md#api-文档](README.md#api-文档)

---

## 📊 文档统计

| 类别 | 文档数 | 总行数(约) |
|------|-------|-----------|
| 快速入门 | 2 | ~1,000 行 |
| 开发文档 | 4 | ~2,000 行 |
| 部署运维 | 2 | ~800 行 |
| 项目状态 | 1 | ~650 行 |
| API 文档 | 3 | 自动生成 |
| **总计** | **12** | **~4,450 行** |

---

## 🔄 文档更新记录

| 日期 | 文档 | 更新内容 |
|------|------|---------|
| 2025-10-23 | PROJECT_STATUS.md | 新增统一的项目状态报告 ✨ |
| 2025-10-23 | DOCS_INDEX.md | 更新文档索引,删除重复文档 |
| 2025-10-22 | 所有文档 | Phase 2.5 全面更新 |
| 2025-10-22 | AUTH.md | 新增认证系统文档 |
| 2025-10-22 | TEST_SUMMARY.md | 新增 Handler 测试报告 |
| 2025-10-22 | DEPLOYMENT.md | 新增部署指南 |
| 2025-10-22 | OPERATIONS.md | 新增运维手册 |

---

## 📞 获取帮助

如有疑问,请通过以下方式获取帮助:

1. **查看文档**: 使用本索引查找相关文档
2. **查看示例**: 参考 [README.md](README.md#api-文档) 中的 API 示例
3. **查看 Swagger**: 访问 http://localhost:8080/swagger/index.html
4. **提交 Issue**: GitHub Issues (如果项目已托管)

---

**祝您使用愉快！**

---

**文档维护者**: OpAMP Platform 开发团队
**文档版本**: v1.1
**生成时间**: 2025-10-23
