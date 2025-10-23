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
	s.logger.Info(fmt.Sprintf("Package created: %s-%s-%s-%s",
		pkg.Name, pkg.Version, pkg.Platform, pkg.Arch))
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
	if err := s.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Find(&packages).Error; err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	return packages, nil
}

// GetLatestPackage 获取最新版本的软件包
func (s *Store) GetLatestPackage(ctx context.Context, name, platform, arch string) (*model.Package, error) {
	var pkg model.Package
	err := s.db.WithContext(ctx).
		Where("name = ? AND platform = ? AND arch = ? AND is_active = ?", name, platform, arch, true).
		Order("version DESC, created_at DESC").
		First(&pkg).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get latest package: %w", err)
	}
	return &pkg, nil
}

// UpdatePackage 更新软件包
func (s *Store) UpdatePackage(ctx context.Context, pkg *model.Package) error {
	if err := s.db.WithContext(ctx).Save(pkg).Error; err != nil {
		return fmt.Errorf("failed to update package: %w", err)
	}
	return nil
}

// DeletePackage 删除软件包
func (s *Store) DeletePackage(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&model.Package{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete package: %w", err)
	}
	s.logger.Info(fmt.Sprintf("Package deleted: ID %d", id))
	return nil
}
