package packagemgr

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
	store   PackageStore
	storage FileStorage
	logger  *zap.Logger
}

// NewManager 创建包管理器
func NewManager(store PackageStore, storage FileStorage, logger *zap.Logger) *Manager {
	return &Manager{
		store:   store,
		storage: storage,
		logger:  logger,
	}
}

// NewManagerWithConcreteTypes 使用具体类型创建包管理器(保持向后兼容)
func NewManagerWithConcreteTypes(store *postgres.Store, storage *storage.MinIOClient, logger *zap.Logger) *Manager {
	return NewManager(store, storage, logger)
}

// UploadPackage 上传软件包
func (m *Manager) UploadPackage(ctx context.Context, pkg *model.Package, reader io.Reader) error {
	// 生成存储路径
	objectName := fmt.Sprintf("packages/%s/%s/%s-%s/%s",
		pkg.Name, pkg.Version, pkg.Platform, pkg.Arch, pkg.Name)

	// 读取文件内容并计算哈希
	hash := sha256.New()
	teeReader := io.TeeReader(reader, hash)

	// 上传到 MinIO
	if err := m.storage.UploadFile(ctx, objectName, teeReader, pkg.FileSize, "application/octet-stream"); err != nil {
		return fmt.Errorf("failed to upload package: %w", err)
	}

	// 计算校验和
	pkg.Checksum = hex.EncodeToString(hash.Sum(nil))
	pkg.StoragePath = objectName
	pkg.IsActive = true

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

// ListPackages 列出所有软件包
func (m *Manager) ListPackages(ctx context.Context) ([]*model.Package, error) {
	return m.store.ListPackages(ctx)
}

// GetPackage 获取单个软件包
func (m *Manager) GetPackage(ctx context.Context, id uint) (*model.Package, error) {
	return m.store.GetPackage(ctx, id)
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
