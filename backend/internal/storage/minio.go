package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// MinIOClient MinIO 客户端封装
type MinIOClient struct {
	client *minio.Client
	bucket string
	logger *zap.Logger
}

// Config MinIO 配置
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// NewMinIOClient 创建 MinIO 客户端
func NewMinIOClient(config Config, logger *zap.Logger) (*MinIOClient, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// 确保 bucket 存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		logger.Info("MinIO bucket created", zap.String("bucket", config.Bucket))
	}

	logger.Info("MinIO client initialized",
		zap.String("endpoint", config.Endpoint),
		zap.String("bucket", config.Bucket))

	return &MinIOClient{
		client: client,
		bucket: config.Bucket,
		logger: logger,
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

	m.logger.Info("File uploaded to MinIO",
		zap.String("object", objectName),
		zap.Int64("size", size))

	return nil
}

// DownloadFile 下载文件
func (m *MinIOClient) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	// 验证对象是否存在
	_, err = object.Stat()
	if err != nil {
		object.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return object, nil
}

// DeleteFile 删除文件
func (m *MinIOClient) DeleteFile(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	m.logger.Info("File deleted from MinIO", zap.String("object", objectName))

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

// FileExists 检查文件是否存在
func (m *MinIOClient) FileExists(ctx context.Context, objectName string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
