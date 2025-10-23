package packagemgr

import (
	"context"
	"io"

	"github.com/cc1024201/opamp-platform/internal/model"
)

// PackageStore 定义包数据库存储接口
type PackageStore interface {
	CreatePackage(ctx context.Context, pkg *model.Package) error
	GetPackage(ctx context.Context, id uint) (*model.Package, error)
	ListPackages(ctx context.Context) ([]*model.Package, error)
	DeletePackage(ctx context.Context, id uint) error
	GetLatestPackage(ctx context.Context, name, platform, arch string) (*model.Package, error)
}

// FileStorage 定义文件存储接口
type FileStorage interface {
	UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, objectName string) error
}
