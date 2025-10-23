package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware Prometheus 监控中间件
func PrometheusMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 获取请求大小
		requestSize := computeRequestSize(c.Request)

		// 处理请求
		c.Next()

		// 计算请求耗时
		duration := time.Since(start).Seconds()

		// 获取响应信息
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()

		// 如果路径为空（如 404），使用请求的 URL 路径
		if path == "" {
			path = c.Request.URL.Path
		}

		// 记录 HTTP 请求指标
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
		metrics.HTTPRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
		metrics.HTTPResponseSize.WithLabelValues(method, path).Observe(float64(c.Writer.Size()))
	}
}

// computeRequestSize 计算请求大小
func computeRequestSize(r any) int64 {
	// 简化实现，实际可以根据需要更精确
	return 0
}
