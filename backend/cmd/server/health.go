package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthStatus 健康检查状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Latency string       `json:"latency,omitempty"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  int64                      `json:"timestamp"`
	Version    string                     `json:"version"`
	Components map[string]ComponentHealth `json:"components"`
}

// healthCheckHandler 详细的健康检查
func healthCheckHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		response := HealthResponse{
			Status:     HealthStatusHealthy,
			Timestamp:  time.Now().Unix(),
			Version:    "1.0.0",
			Components: make(map[string]ComponentHealth),
		}

		// 检查数据库
		dbHealth := checkDatabase(ctx, db)
		response.Components["database"] = dbHealth
		if dbHealth.Status == HealthStatusUnhealthy {
			response.Status = HealthStatusUnhealthy
		} else if dbHealth.Status == HealthStatusDegraded && response.Status == HealthStatusHealthy {
			response.Status = HealthStatusDegraded
		}

		// 检查 OpAMP 服务（基本检查）
		opampHealth := ComponentHealth{
			Status:  HealthStatusHealthy,
			Message: "OpAMP server is running",
		}
		response.Components["opamp"] = opampHealth

		// 设置 HTTP 状态码
		statusCode := http.StatusOK
		if response.Status == HealthStatusDegraded {
			statusCode = http.StatusOK // 降级状态仍然返回 200
		} else if response.Status == HealthStatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, response)
	}
}

// checkDatabase 检查数据库连接
func checkDatabase(ctx context.Context, db *gorm.DB) ComponentHealth {
	start := time.Now()

	sqlDB, err := db.DB()
	if err != nil {
		return ComponentHealth{
			Status:  HealthStatusUnhealthy,
			Message: "failed to get database connection: " + err.Error(),
		}
	}

	// Ping 数据库
	if err := sqlDB.PingContext(ctx); err != nil {
		return ComponentHealth{
			Status:  HealthStatusUnhealthy,
			Message: "database ping failed: " + err.Error(),
		}
	}

	latency := time.Since(start)

	// 检查连接池状态
	stats := sqlDB.Stats()
	if stats.OpenConnections >= stats.MaxOpenConnections {
		return ComponentHealth{
			Status:  HealthStatusDegraded,
			Message: "database connection pool exhausted",
			Latency: latency.String(),
		}
	}

	return ComponentHealth{
		Status:  HealthStatusHealthy,
		Message: "database connection successful",
		Latency: latency.String(),
	}
}

// readinessHandler Kubernetes 就绪探针
func readinessHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// 检查数据库
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(ctx) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"ready": false,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ready": true,
		})
	}
}

// livenessHandler Kubernetes 存活探针
func livenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"alive": true,
		})
	}
}
