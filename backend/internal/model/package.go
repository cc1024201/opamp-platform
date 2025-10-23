package model

import (
	"time"
)

// Package Agent 软件包
type Package struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex:idx_package_unique;size:255;not null" json:"name"`
	Version     string    `gorm:"uniqueIndex:idx_package_unique;size:50;not null" json:"version"`
	Platform    string    `gorm:"uniqueIndex:idx_package_unique;size:50;not null" json:"platform"` // linux/windows/darwin
	Arch        string    `gorm:"uniqueIndex:idx_package_unique;size:50;not null" json:"arch"`     // amd64/arm64/386
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
