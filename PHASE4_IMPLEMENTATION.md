# Phase 4 实施计划 - 核心功能完善

**版本**: v2.1.0
**开始日期**: 2025-10-23
**预计完成**: 2025-11-23 (1个月)
**负责人**: Development Team

---

## 📋 总体目标

完成 OpAMP 协议的完整实现和前端核心页面开发,使系统达到可生产使用的状态。

### 关键成果
- ✅ Agent 可以自动更新软件包
- ✅ 配置可以实时推送到 Agent
- ✅ 用户可以通过界面查看 Agent 详情
- ✅ 用户可以通过界面编辑配置
- ✅ 仪表盘显示实时统计数据

---

## 🔧 Part 1: 后端 OpAMP 协议完善

### 任务 1.1: Agent 包管理系统 ⭐⭐⭐

**预计时间**: 5 天
**优先级**: 高

#### 功能需求
1. Agent 软件包上传和存储
2. 多平台支持 (Linux/Windows/macOS)
3. 版本管理和发布
4. Agent 自动更新机制
5. 下载进度跟踪

#### 实现步骤

##### 1.1.1 数据模型设计

**文件**: `internal/model/package.go`

```go
package model

import (
    "time"
)

// Package Agent 软件包
type Package struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    Name        string    `gorm:"uniqueIndex;size:255;not null" json:"name"`
    Version     string    `gorm:"index;size:50;not null" json:"version"`
    Platform    string    `gorm:"index;size:50;not null" json:"platform"` // linux/windows/darwin
    Arch        string    `gorm:"index;size:50;not null" json:"arch"`     // amd64/arm64/386
    FileSize    int64     `json:"file_size"`
    Checksum    string    `gorm:"size:64" json:"checksum"`                // SHA256
    StoragePath string    `gorm:"size:500" json:"storage_path"`           // MinIO 路径
    Description string    `gorm:"type:text" json:"description"`
    IsActive    bool      `gorm:"default:true" json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Package) TableName() string {
    return "packages"
}
```

**迁移文件**: `backend/migrations/000003_add_packages.up.sql`

```sql
CREATE TABLE IF NOT EXISTS packages (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    arch VARCHAR(50) NOT NULL,
    file_size BIGINT,
    checksum VARCHAR(64),
    storage_path VARCHAR(500),
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_packages_name_version_platform_arch
    ON packages(name, version, platform, arch);

CREATE INDEX idx_packages_platform ON packages(platform);
CREATE INDEX idx_packages_version ON packages(version);
```

##### 1.1.2 存储层实现

**文件**: `internal/store/postgres/package.go`

```go
package postgres

import (
    "context"
    "fmt"

    "github.com/cc1024201/opamp-platform/internal/model"
)

// CreatePackage 创建软件包
func (s *Store) CreatePackage(ctx context.Context, pkg *model.Package) error {
    if err := s.db.WithContext(ctx).Create(pkg).Error; err != nil {
        return fmt.Errorf("failed to create package: %w", err)
    }
    return nil
}

// GetPackage 获取软件包
func (s *Store) GetPackage(ctx context.Context, id uint) (*model.Package, error) {
    var pkg model.Package
    if err := s.db.WithContext(ctx).First(&pkg, id).Error; err != nil {
        return nil, fmt.Errorf("failed to get package: %w", err)
    }
    return &pkg, nil
}

// GetPackageByVersion 根据版本获取软件包
func (s *Store) GetPackageByVersion(ctx context.Context, name, version, platform, arch string) (*model.Package, error) {
    var pkg model.Package
    err := s.db.WithContext(ctx).
        Where("name = ? AND version = ? AND platform = ? AND arch = ?", name, version, platform, arch).
        First(&pkg).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get package: %w", err)
    }
    return &pkg, nil
}

// ListPackages 列出所有软件包
func (s *Store) ListPackages(ctx context.Context) ([]*model.Package, error) {
    var packages []*model.Package
    if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&packages).Error; err != nil {
        return nil, fmt.Errorf("failed to list packages: %w", err)
    }
    return packages, nil
}

// GetLatestPackage 获取最新版本的软件包
func (s *Store) GetLatestPackage(ctx context.Context, name, platform, arch string) (*model.Package, error) {
    var pkg model.Package
    err := s.db.WithContext(ctx).
        Where("name = ? AND platform = ? AND arch = ? AND is_active = ?", name, platform, arch, true).
        Order("version DESC").
        First(&pkg).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get latest package: %w", err)
    }
    return &pkg, nil
}

// DeletePackage 删除软件包
func (s *Store) DeletePackage(ctx context.Context, id uint) error {
    if err := s.db.WithContext(ctx).Delete(&model.Package{}, id).Error; err != nil {
        return fmt.Errorf("failed to delete package: %w", err)
    }
    return nil
}
```

##### 1.1.3 MinIO 文件存储

**文件**: `internal/storage/minio.go`

```go
package storage

import (
    "context"
    "fmt"
    "io"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient MinIO 客户端封装
type MinIOClient struct {
    client *minio.Client
    bucket string
}

// NewMinIOClient 创建 MinIO 客户端
func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOClient, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create minio client: %w", err)
    }

    // 确保 bucket 存在
    ctx := context.Background()
    exists, err := client.BucketExists(ctx, bucket)
    if err != nil {
        return nil, fmt.Errorf("failed to check bucket: %w", err)
    }
    if !exists {
        if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
            return nil, fmt.Errorf("failed to create bucket: %w", err)
        }
    }

    return &MinIOClient{
        client: client,
        bucket: bucket,
    }, nil
}

// UploadFile 上传文件
func (m *MinIOClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
    _, err := m.client.PutObject(ctx, m.bucket, objectName, reader, size, minio.PutObjectOptions{
        ContentType: contentType,
    })
    if err != nil {
        return fmt.Errorf("failed to upload file: %w", err)
    }
    return nil
}

// DownloadFile 下载文件
func (m *MinIOClient) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
    object, err := m.client.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to download file: %w", err)
    }
    return object, nil
}

// DeleteFile 删除文件
func (m *MinIOClient) DeleteFile(ctx context.Context, objectName string) error {
    err := m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{})
    if err != nil {
        return fmt.Errorf("failed to delete file: %w", err)
    }
    return nil
}

// GetFileInfo 获取文件信息
func (m *MinIOClient) GetFileInfo(ctx context.Context, objectName string) (*minio.ObjectInfo, error) {
    info, err := m.client.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to get file info: %w", err)
    }
    return &info, nil
}
```

##### 1.1.4 包管理器

**文件**: `internal/package/manager.go`

```go
package packagemanager

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"

    "github.com/cc1024201/opamp-platform/internal/model"
    "github.com/cc1024201/opamp-platform/internal/storage"
    "github.com/cc1024201/opamp-platform/internal/store/postgres"
    "go.uber.org/zap"
)

// Manager 包管理器
type Manager struct {
    store   *postgres.Store
    storage *storage.MinIOClient
    logger  *zap.Logger
}

// NewManager 创建包管理器
func NewManager(store *postgres.Store, storage *storage.MinIOClient, logger *zap.Logger) *Manager {
    return &Manager{
        store:   store,
        storage: storage,
        logger:  logger,
    }
}

// UploadPackage 上传软件包
func (m *Manager) UploadPackage(ctx context.Context, pkg *model.Package, reader io.Reader) error {
    // 计算文件哈希
    hash := sha256.New()
    teeReader := io.TeeReader(reader, hash)

    // 生成存储路径
    objectName := fmt.Sprintf("packages/%s/%s/%s-%s/%s",
        pkg.Name, pkg.Version, pkg.Platform, pkg.Arch, pkg.Name)

    // 上传到 MinIO
    if err := m.storage.UploadFile(ctx, objectName, teeReader, pkg.FileSize, "application/octet-stream"); err != nil {
        return fmt.Errorf("failed to upload package: %w", err)
    }

    // 计算校验和
    pkg.Checksum = hex.EncodeToString(hash.Sum(nil))
    pkg.StoragePath = objectName

    // 保存到数据库
    if err := m.store.CreatePackage(ctx, pkg); err != nil {
        // 如果数据库保存失败,删除已上传的文件
        _ = m.storage.DeleteFile(ctx, objectName)
        return fmt.Errorf("failed to save package metadata: %w", err)
    }

    m.logger.Info("Package uploaded successfully",
        zap.String("name", pkg.Name),
        zap.String("version", pkg.Version),
        zap.String("platform", pkg.Platform),
        zap.String("arch", pkg.Arch))

    return nil
}

// DownloadPackage 下载软件包
func (m *Manager) DownloadPackage(ctx context.Context, id uint) (io.ReadCloser, *model.Package, error) {
    pkg, err := m.store.GetPackage(ctx, id)
    if err != nil {
        return nil, nil, err
    }

    reader, err := m.storage.DownloadFile(ctx, pkg.StoragePath)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to download package: %w", err)
    }

    return reader, pkg, nil
}

// GetLatestVersion 获取最新版本
func (m *Manager) GetLatestVersion(ctx context.Context, name, platform, arch string) (*model.Package, error) {
    return m.store.GetLatestPackage(ctx, name, platform, arch)
}

// DeletePackage 删除软件包
func (m *Manager) DeletePackage(ctx context.Context, id uint) error {
    pkg, err := m.store.GetPackage(ctx, id)
    if err != nil {
        return err
    }

    // 删除文件
    if err := m.storage.DeleteFile(ctx, pkg.StoragePath); err != nil {
        m.logger.Error("Failed to delete package file", zap.Error(err))
        // 继续删除数据库记录
    }

    // 删除数据库记录
    return m.store.DeletePackage(ctx, id)
}
```

##### 1.1.5 API 处理器

**文件**: `cmd/server/package_handlers.go`

```go
package main

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    packagemanager "github.com/cc1024201/opamp-platform/internal/package"
)

// uploadPackageHandler 上传软件包
// @Summary      上传软件包
// @Description  上传新的 Agent 软件包
// @Tags         packages
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "软件包文件"
// @Param        name formData string true "包名称"
// @Param        version formData string true "版本号"
// @Param        platform formData string true "平台 (linux/windows/darwin)"
// @Param        arch formData string true "架构 (amd64/arm64/386)"
// @Param        description formData string false "描述"
// @Success      200 {object} model.Package
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Security     BearerAuth
// @Router       /packages [post]
func uploadPackageHandler(pm *packagemanager.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 解析表单
        file, header, err := c.Request.FormFile("file")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
            return
        }
        defer file.Close()

        pkg := &model.Package{
            Name:        c.PostForm("name"),
            Version:     c.PostForm("version"),
            Platform:    c.PostForm("platform"),
            Arch:        c.PostForm("arch"),
            Description: c.PostForm("description"),
            FileSize:    header.Size,
            IsActive:    true,
        }

        // 验证必填字段
        if pkg.Name == "" || pkg.Version == "" || pkg.Platform == "" || pkg.Arch == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "name, version, platform and arch are required"})
            return
        }

        // 上传包
        if err := pm.UploadPackage(c.Request.Context(), pkg, file); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, pkg)
    }
}

// listPackagesHandler 列出所有软件包
// @Summary      列出软件包
// @Description  获取所有软件包列表
// @Tags         packages
// @Produce      json
// @Success      200 {array} model.Package
// @Failure      500 {object} map[string]string
// @Security     BearerAuth
// @Router       /packages [get]
func listPackagesHandler(pm *packagemanager.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        packages, err := pm.store.ListPackages(c.Request.Context())
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, packages)
    }
}

// downloadPackageHandler 下载软件包
// @Summary      下载软件包
// @Description  下载指定的软件包文件
// @Tags         packages
// @Produce      application/octet-stream
// @Param        id path int true "包 ID"
// @Success      200 {file} binary
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Security     BearerAuth
// @Router       /packages/{id}/download [get]
func downloadPackageHandler(pm *packagemanager.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid package id"})
            return
        }

        reader, pkg, err := pm.DownloadPackage(c.Request.Context(), uint(id))
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        defer reader.Close()

        c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s-%s-%s",
            pkg.Name, pkg.Version, pkg.Platform, pkg.Arch))
        c.Header("Content-Type", "application/octet-stream")
        c.Header("Content-Length", strconv.FormatInt(pkg.FileSize, 10))

        io.Copy(c.Writer, reader)
    }
}

// deletePackageHandler 删除软件包
// @Summary      删除软件包
// @Description  删除指定的软件包
// @Tags         packages
// @Produce      json
// @Param        id path int true "包 ID"
// @Success      200 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Security     BearerAuth
// @Router       /packages/{id} [delete]
func deletePackageHandler(pm *packagemanager.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid package id"})
            return
        }

        if err := pm.DeletePackage(c.Request.Context(), uint(id)); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "package deleted successfully"})
    }
}
```

##### 1.1.6 集成到 main.go

在 `cmd/server/main.go` 中添加:

```go
import (
    packagemanager "github.com/cc1024201/opamp-platform/internal/package"
    "github.com/cc1024201/opamp-platform/internal/storage"
)

// 在 main 函数中初始化
minioClient, err := storage.NewMinIOClient(
    viper.GetString("minio.endpoint"),
    viper.GetString("minio.access_key"),
    viper.GetString("minio.secret_key"),
    viper.GetString("minio.bucket"),
    viper.GetBool("minio.use_ssl"),
)
if err != nil {
    logger.Fatal("Failed to initialize MinIO client", zap.Error(err))
}

packageManager := packagemanager.NewManager(store, minioClient, logger)

// 添加路由
packages := authenticated.Group("/packages")
{
    packages.GET("", listPackagesHandler(packageManager))
    packages.POST("", uploadPackageHandler(packageManager))
    packages.GET("/:id/download", downloadPackageHandler(packageManager))
    packages.DELETE("/:id", deletePackageHandler(packageManager))
}
```

#### 测试要求

**文件**: `internal/package/manager_test.go`

```go
package packagemanager

import (
    "bytes"
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUploadPackage(t *testing.T) {
    // TODO: 实现测试
    t.Skip("需要集成测试环境")
}

func TestDownloadPackage(t *testing.T) {
    // TODO: 实现测试
    t.Skip("需要集成测试环境")
}

func TestGetLatestVersion(t *testing.T) {
    // TODO: 实现测试
    t.Skip("需要集成测试环境")
}
```

---

### 任务 1.2: 配置热更新机制 ⭐⭐⭐

**预计时间**: 3 天
**优先级**: 高

#### 功能需求
1. 配置变更检测
2. OpAMP 消息推送
3. Agent 配置应用确认
4. 配置应用状态跟踪

#### 实现步骤

##### 1.2.1 增强 OpAMP 服务器

**文件**: `internal/opamp/config_update.go`

```go
package opamp

import (
    "context"
    "fmt"

    "github.com/open-telemetry/opamp-go/protobufs"
    "go.uber.org/zap"
)

// SendConfigUpdate 发送配置更新到 Agent
func (s *Server) SendConfigUpdate(agentID string, config *protobufs.AgentRemoteConfig) error {
    s.mu.RLock()
    conn, exists := s.connections[agentID]
    s.mu.RUnlock()

    if !exists {
        return fmt.Errorf("agent not connected: %s", agentID)
    }

    // 发送配置更新消息
    err := conn.Send(context.Background(), &protobufs.ServerToAgent{
        RemoteConfig: config,
    })

    if err != nil {
        s.logger.Error("Failed to send config update",
            zap.String("agent_id", agentID),
            zap.Error(err))
        return err
    }

    s.logger.Info("Config update sent",
        zap.String("agent_id", agentID))

    return nil
}

// BroadcastConfigUpdate 广播配置更新到多个 Agents
func (s *Server) BroadcastConfigUpdate(agentIDs []string, config *protobufs.AgentRemoteConfig) error {
    var failedAgents []string

    for _, agentID := range agentIDs {
        if err := s.SendConfigUpdate(agentID, config); err != nil {
            failedAgents = append(failedAgents, agentID)
        }
    }

    if len(failedAgents) > 0 {
        return fmt.Errorf("failed to send config to %d agents: %v",
            len(failedAgents), failedAgents)
    }

    return nil
}
```

##### 1.2.2 配置应用状态跟踪

**文件**: `internal/model/config_deployment.go`

```go
package model

import "time"

// ConfigDeployment 配置部署记录
type ConfigDeployment struct {
    ID              uint      `gorm:"primaryKey" json:"id"`
    ConfigurationID uint      `gorm:"index;not null" json:"configuration_id"`
    AgentID         uint      `gorm:"index;not null" json:"agent_id"`
    Status          string    `gorm:"size:50;not null" json:"status"` // pending/applied/failed
    Version         string    `gorm:"size:100" json:"version"`
    Error           string    `gorm:"type:text" json:"error,omitempty"`
    AppliedAt       *time.Time `json:"applied_at,omitempty"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`

    Configuration *Configuration `gorm:"foreignKey:ConfigurationID" json:"configuration,omitempty"`
    Agent         *Agent         `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

func (ConfigDeployment) TableName() string {
    return "config_deployments"
}
```

#### 测试要求
- 单元测试: 配置推送逻辑
- 集成测试: 模拟 Agent 接收配置
- 状态跟踪测试

---

### 任务 1.3: Agent 状态管理增强 ⭐⭐

**预计时间**: 2 天
**优先级**: 中

#### 功能需求
1. Agent 心跳监控
2. 离线检测
3. 连接状态持久化
4. Agent 元数据更新

**实现细节见 ROADMAP.md**

---

## 🎨 Part 2: 前端核心页面开发

### 任务 2.1: Agent 详情页面 ⭐⭐⭐

**预计时间**: 3 天
**优先级**: 高

#### 页面设计

**文件**: `frontend/src/pages/agents/AgentDetailPage.tsx`

```typescript
import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Grid,
  Chip,
  Button,
  Tab,
  Tabs,
  Table,
  TableBody,
  TableCell,
  TableRow,
  Alert,
  CircularProgress,
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Delete as DeleteIcon,
  Settings as SettingsIcon,
} from '@mui/icons-material';
import { agentApi } from '@/services/api';
import { Agent } from '@/types/agent';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;
  return (
    <div hidden={value !== index} {...other}>
      {value === index && <Box sx={{ pt: 3 }}>{children}</Box>}
    </div>
  );
}

export default function AgentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [agent, setAgent] = useState<Agent | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tabValue, setTabValue] = useState(0);

  useEffect(() => {
    loadAgent();
  }, [id]);

  const loadAgent = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await agentApi.getAgent(Number(id));
      setAgent(data);
    } catch (err: any) {
      setError(err.message || '加载 Agent 失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!window.confirm('确定要删除这个 Agent 吗?')) return;

    try {
      await agentApi.deleteAgent(Number(id));
      navigate('/agents');
    } catch (err: any) {
      alert('删除失败: ' + err.message);
    }
  };

  const handleTabChange = (_: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  if (error || !agent) {
    return (
      <Box>
        <Alert severity="error">{error || 'Agent 不存在'}</Alert>
        <Button onClick={() => navigate('/agents')} sx={{ mt: 2 }}>
          返回列表
        </Button>
      </Box>
    );
  }

  return (
    <Box>
      {/* 页头 */}
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Agent 详情</Typography>
        <Box>
          <Button
            startIcon={<RefreshIcon />}
            onClick={loadAgent}
            sx={{ mr: 1 }}
          >
            刷新
          </Button>
          <Button
            startIcon={<SettingsIcon />}
            variant="outlined"
            sx={{ mr: 1 }}
          >
            配置
          </Button>
          <Button
            startIcon={<DeleteIcon />}
            variant="outlined"
            color="error"
            onClick={handleDelete}
          >
            删除
          </Button>
        </Box>
      </Box>

      {/* 基本信息卡片 */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" color="text.secondary">
                Agent ID
              </Typography>
              <Typography variant="body1" sx={{ mt: 0.5 }}>
                {agent.agent_id}
              </Typography>
            </Grid>

            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" color="text.secondary">
                状态
              </Typography>
              <Chip
                label={agent.status}
                color={agent.status === 'connected' ? 'success' : 'default'}
                sx={{ mt: 0.5 }}
              />
            </Grid>

            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" color="text.secondary">
                版本
              </Typography>
              <Typography variant="body1" sx={{ mt: 0.5 }}>
                {agent.version || 'N/A'}
              </Typography>
            </Grid>

            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" color="text.secondary">
                主机名
              </Typography>
              <Typography variant="body1" sx={{ mt: 0.5 }}>
                {agent.hostname || 'N/A'}
              </Typography>
            </Grid>

            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" color="text.secondary">
                最后连接时间
              </Typography>
              <Typography variant="body1" sx={{ mt: 0.5 }}>
                {new Date(agent.last_seen).toLocaleString()}
              </Typography>
            </Grid>

            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" color="text.secondary">
                创建时间
              </Typography>
              <Typography variant="body1" sx={{ mt: 0.5 }}>
                {new Date(agent.created_at).toLocaleString()}
              </Typography>
            </Grid>

            {agent.labels && Object.keys(agent.labels).length > 0 && (
              <Grid item xs={12}>
                <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 1 }}>
                  标签
                </Typography>
                <Box>
                  {Object.entries(agent.labels).map(([key, value]) => (
                    <Chip
                      key={key}
                      label={`${key}: ${value}`}
                      size="small"
                      sx={{ mr: 1, mb: 1 }}
                    />
                  ))}
                </Box>
              </Grid>
            )}
          </Grid>
        </CardContent>
      </Card>

      {/* Tabs */}
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={tabValue} onChange={handleTabChange}>
          <Tab label="系统信息" />
          <Tab label="配置历史" />
          <Tab label="性能指标" />
        </Tabs>
      </Box>

      <TabPanel value={tabValue} index={0}>
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              系统信息
            </Typography>
            <Table>
              <TableBody>
                <TableRow>
                  <TableCell>操作系统</TableCell>
                  <TableCell>{agent.os || 'N/A'}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>架构</TableCell>
                  <TableCell>{agent.arch || 'N/A'}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell>IP 地址</TableCell>
                  <TableCell>{agent.ip_address || 'N/A'}</TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </TabPanel>

      <TabPanel value={tabValue} index={1}>
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              配置历史
            </Typography>
            <Typography color="text.secondary">
              暂无配置历史记录
            </Typography>
          </CardContent>
        </Card>
      </TabPanel>

      <TabPanel value={tabValue} index={2}>
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              性能指标
            </Typography>
            <Typography color="text.secondary">
              性能监控功能即将推出
            </Typography>
          </CardContent>
        </Card>
      </TabPanel>
    </Box>
  );
}
```

#### API 服务

**文件**: `frontend/src/services/api/agent.ts`

```typescript
import axios from './axios';
import { Agent } from '@/types/agent';

export const agentApi = {
  // 获取 Agent 列表
  async getAgents(): Promise<Agent[]> {
    const response = await axios.get('/agents');
    return response.data;
  },

  // 获取单个 Agent
  async getAgent(id: number): Promise<Agent> {
    const response = await axios.get(`/agents/${id}`);
    return response.data;
  },

  // 删除 Agent
  async deleteAgent(id: number): Promise<void> {
    await axios.delete(`/agents/${id}`);
  },
};
```

---

### 任务 2.2: Configuration 编辑页面 ⭐⭐⭐

**预计时间**: 4 天
**优先级**: 高

#### 集成 Monaco Editor

**安装依赖**:
```bash
cd frontend
npm install @monaco-editor/react monaco-editor
```

**页面实现**: 见 ROADMAP.md 详细设计

---

### 任务 2.3: 仪表盘增强 ⭐⭐

**预计时间**: 2 天
**优先级**: 中

**功能**:
- 实时统计卡片
- 状态分布饼图
- 活动时间线

---

## ✅ 验收标准

### 后端验收
- [ ] 所有 API 测试通过
- [ ] 测试覆盖率 > 80%
- [ ] 可以上传和下载软件包
- [ ] 配置可以推送到 Agent
- [ ] Agent 心跳监控正常

### 前端验收
- [ ] Agent 详情页面功能完整
- [ ] Configuration 编辑器可用
- [ ] 仪表盘图表正常显示
- [ ] 无明显 UI 错误

### 集成测试
- [ ] 端到端测试通过
- [ ] 真实 Agent 连接测试
- [ ] 配置更新测试

---

## 📅 时间表

| 周次 | 任务 | 负责人 | 状态 |
|------|------|--------|------|
| Week 1 | 任务 1.1: Agent 包管理 | Backend Team | 进行中 |
| Week 1-2 | 任务 2.1: Agent 详情页面 | Frontend Team | 待开始 |
| Week 2 | 任务 1.2: 配置热更新 | Backend Team | 待开始 |
| Week 2-3 | 任务 2.2: Configuration 编辑 | Frontend Team | 待开始 |
| Week 3 | 任务 1.3: 状态管理 | Backend Team | 待开始 |
| Week 3 | 任务 2.3: 仪表盘增强 | Frontend Team | 待开始 |
| Week 4 | 集成测试和优化 | All | 待开始 |

---

## 📝 注意事项

1. **代码规范**: 遵循现有的代码风格和命名规范
2. **测试先行**: 重要功能先写测试用例
3. **文档更新**: 及时更新 API 文档和使用文档
4. **Code Review**: 所有代码必须经过审查
5. **版本控制**: 使用 Git 分支管理,功能完成后合并到 main

---

**下一步**: 开始实现任务 1.1 - Agent 包管理系统
