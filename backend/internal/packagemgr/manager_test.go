package packagemgr

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockPackageStore 模拟数据库存储
type MockPackageStore struct {
	mock.Mock
}

func (m *MockPackageStore) CreatePackage(ctx context.Context, pkg *model.Package) error {
	args := m.Called(ctx, pkg)
	return args.Error(0)
}

func (m *MockPackageStore) GetPackage(ctx context.Context, id uint) (*model.Package, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Package), args.Error(1)
}

func (m *MockPackageStore) ListPackages(ctx context.Context) ([]*model.Package, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Package), args.Error(1)
}

func (m *MockPackageStore) DeletePackage(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPackageStore) GetLatestPackage(ctx context.Context, name, platform, arch string) (*model.Package, error) {
	args := m.Called(ctx, name, platform, arch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Package), args.Error(1)
}

// MockFileStorage 模拟 MinIO 存储
type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	args := m.Called(ctx, objectName, reader, size, contentType)
	return args.Error(0)
}

func (m *MockFileStorage) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	args := m.Called(ctx, objectName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockFileStorage) DeleteFile(ctx context.Context, objectName string) error {
	args := m.Called(ctx, objectName)
	return args.Error(0)
}

func TestNewManager(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockStorage, logger)

	assert.NotNil(t, manager)
	assert.Equal(t, mockStore, manager.store)
	assert.Equal(t, mockStorage, manager.storage)
	assert.Equal(t, logger, manager.logger)
}

func TestUploadPackage_Success(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	pkg := &model.Package{
		Name:     "test-agent",
		Version:  "1.0.0",
		Platform: "linux",
		Arch:     "amd64",
		FileSize: 1024,
	}

	reader := strings.NewReader("test file content")
	ctx := context.Background()

	mockStorage.On("UploadFile", ctx, mock.Anything, mock.Anything, pkg.FileSize, "application/octet-stream").Return(nil)
	mockStore.On("CreatePackage", ctx, pkg).Return(nil)

	err := manager.UploadPackage(ctx, pkg, reader)

	assert.NoError(t, err)
	assert.NotEmpty(t, pkg.Checksum)
	assert.NotEmpty(t, pkg.StoragePath)
	assert.True(t, pkg.IsActive)
	mockStorage.AssertExpectations(t)
	mockStore.AssertExpectations(t)
}

func TestUploadPackage_StorageError(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	pkg := &model.Package{
		Name:     "test-agent",
		Version:  "1.0.0",
		Platform: "linux",
		Arch:     "amd64",
		FileSize: 1024,
	}

	reader := strings.NewReader("test file content")
	ctx := context.Background()

	mockStorage.On("UploadFile", ctx, mock.Anything, mock.Anything, pkg.FileSize, "application/octet-stream").
		Return(errors.New("storage error"))

	err := manager.UploadPackage(ctx, pkg, reader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upload package")
	mockStorage.AssertExpectations(t)
	mockStore.AssertNotCalled(t, "CreatePackage")
}

func TestUploadPackage_DatabaseError(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	pkg := &model.Package{
		Name:     "test-agent",
		Version:  "1.0.0",
		Platform: "linux",
		Arch:     "amd64",
		FileSize: 1024,
	}

	reader := strings.NewReader("test file content")
	ctx := context.Background()

	mockStorage.On("UploadFile", ctx, mock.Anything, mock.Anything, pkg.FileSize, "application/octet-stream").Return(nil)
	mockStorage.On("DeleteFile", ctx, mock.Anything).Return(nil)
	mockStore.On("CreatePackage", ctx, pkg).Return(errors.New("database error"))

	err := manager.UploadPackage(ctx, pkg, reader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save package metadata")
	mockStorage.AssertExpectations(t)
	mockStore.AssertExpectations(t)
}

func TestDownloadPackage_Success(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	expectedPkg := &model.Package{
		ID:          1,
		Name:        "test-agent",
		Version:     "1.0.0",
		StoragePath: "packages/test-agent/1.0.0/linux-amd64/test-agent",
	}

	ctx := context.Background()
	mockReader := io.NopCloser(strings.NewReader("file content"))

	mockStore.On("GetPackage", ctx, uint(1)).Return(expectedPkg, nil)
	mockStorage.On("DownloadFile", ctx, expectedPkg.StoragePath).Return(mockReader, nil)

	reader, pkg, err := manager.DownloadPackage(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, reader)
	assert.Equal(t, expectedPkg, pkg)
	mockStore.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestDownloadPackage_PackageNotFound(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	ctx := context.Background()

	mockStore.On("GetPackage", ctx, uint(1)).Return(nil, errors.New("package not found"))

	reader, pkg, err := manager.DownloadPackage(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, reader)
	assert.Nil(t, pkg)
	mockStore.AssertExpectations(t)
	mockStorage.AssertNotCalled(t, "DownloadFile")
}

func TestGetLatestVersion_Success(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	expectedPkg := &model.Package{
		ID:       1,
		Name:     "test-agent",
		Version:  "2.0.0",
		Platform: "linux",
		Arch:     "amd64",
	}

	ctx := context.Background()

	mockStore.On("GetLatestPackage", ctx, "test-agent", "linux", "amd64").Return(expectedPkg, nil)

	pkg, err := manager.GetLatestVersion(ctx, "test-agent", "linux", "amd64")

	assert.NoError(t, err)
	assert.Equal(t, expectedPkg, pkg)
	mockStore.AssertExpectations(t)
}

func TestListPackages_Success(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	expectedPkgs := []*model.Package{
		{ID: 1, Name: "agent-1", Version: "1.0.0"},
		{ID: 2, Name: "agent-2", Version: "2.0.0"},
	}

	ctx := context.Background()

	mockStore.On("ListPackages", ctx).Return(expectedPkgs, nil)

	pkgs, err := manager.ListPackages(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedPkgs, pkgs)
	assert.Len(t, pkgs, 2)
	mockStore.AssertExpectations(t)
}

func TestGetPackage_Success(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	expectedPkg := &model.Package{
		ID:      1,
		Name:    "test-agent",
		Version: "1.0.0",
	}

	ctx := context.Background()

	mockStore.On("GetPackage", ctx, uint(1)).Return(expectedPkg, nil)

	pkg, err := manager.GetPackage(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedPkg, pkg)
	mockStore.AssertExpectations(t)
}

func TestDeletePackage_Success(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	pkg := &model.Package{
		ID:          1,
		Name:        "test-agent",
		StoragePath: "packages/test-agent/1.0.0/linux-amd64/test-agent",
	}

	ctx := context.Background()

	mockStore.On("GetPackage", ctx, uint(1)).Return(pkg, nil)
	mockStorage.On("DeleteFile", ctx, pkg.StoragePath).Return(nil)
	mockStore.On("DeletePackage", ctx, uint(1)).Return(nil)

	err := manager.DeletePackage(ctx, 1)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestDeletePackage_StorageError_ContinuesWithDBDeletion(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	pkg := &model.Package{
		ID:          1,
		Name:        "test-agent",
		StoragePath: "packages/test-agent/1.0.0/linux-amd64/test-agent",
	}

	ctx := context.Background()

	mockStore.On("GetPackage", ctx, uint(1)).Return(pkg, nil)
	mockStorage.On("DeleteFile", ctx, pkg.StoragePath).Return(errors.New("storage error"))
	mockStore.On("DeletePackage", ctx, uint(1)).Return(nil)

	err := manager.DeletePackage(ctx, 1)

	// 即使存储删除失败,数据库删除应该继续,整体操作应该成功
	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestDeletePackage_PackageNotFound(t *testing.T) {
	mockStore := new(MockPackageStore)
	mockStorage := new(MockFileStorage)
	logger := zap.NewNop()

	manager := &Manager{
		store:   mockStore,
		storage: mockStorage,
		logger:  logger,
	}

	ctx := context.Background()

	mockStore.On("GetPackage", ctx, uint(1)).Return(nil, errors.New("package not found"))

	err := manager.DeletePackage(ctx, 1)

	assert.Error(t, err)
	mockStore.AssertExpectations(t)
	mockStorage.AssertNotCalled(t, "DeleteFile")
	mockStore.AssertNotCalled(t, "DeletePackage")
}
