package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/cc1024201/opamp-platform/internal/packagemgr"
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
func uploadPackageHandler(pm *packagemgr.Manager) gin.HandlerFunc {
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
func listPackagesHandler(pm *packagemgr.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		packages, err := pm.ListPackages(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, packages)
	}
}

// getPackageHandler 获取软件包详情
// @Summary      获取软件包详情
// @Description  获取指定软件包的详细信息
// @Tags         packages
// @Produce      json
// @Param        id path int true "包 ID"
// @Success      200 {object} model.Package
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Security     BearerAuth
// @Router       /packages/{id} [get]
func getPackageHandler(pm *packagemgr.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid package id"})
			return
		}

		pkg, err := pm.GetPackage(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, pkg)
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
func downloadPackageHandler(pm *packagemgr.Manager) gin.HandlerFunc {
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

		// 设置响应头
		filename := fmt.Sprintf("%s-%s-%s-%s", pkg.Name, pkg.Version, pkg.Platform, pkg.Arch)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Length", strconv.FormatInt(pkg.FileSize, 10))

		// 流式传输文件
		_, err = io.Copy(c.Writer, reader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to download file"})
			return
		}
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
func deletePackageHandler(pm *packagemgr.Manager) gin.HandlerFunc {
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
