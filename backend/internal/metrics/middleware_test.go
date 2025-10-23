package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "test",
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "test",
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
			},
			[]string{"method", "path"},
		),
		HTTPRequestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "test",
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "test",
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
			},
			[]string{"method", "path"},
		),
	}
}

func TestPrometheusMiddleware(t *testing.T) {
	t.Run("middleware records HTTP metrics", func(t *testing.T) {
		m := createTestMetrics()
		registry := prometheus.NewRegistry()
		registry.MustRegister(m.HTTPRequestsTotal)
		registry.MustRegister(m.HTTPRequestDuration)
		registry.MustRegister(m.HTTPRequestSize)
		registry.MustRegister(m.HTTPResponseSize)

		router := setupTestRouter()
		router.Use(PrometheusMiddleware(m))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 发送请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		// 验证指标被记录
		metrics, err := registry.Gather()
		require.NoError(t, err)

		var foundTotal, foundDuration, foundReqSize, foundRespSize bool

		for _, mf := range metrics {
			switch mf.GetName() {
			case "test_http_requests_total":
				foundTotal = true
				require.Len(t, mf.GetMetric(), 1)
				assert.Equal(t, 1.0, mf.GetMetric()[0].GetCounter().GetValue())
			case "test_http_request_duration_seconds":
				foundDuration = true
				require.Len(t, mf.GetMetric(), 1)
				// 验证有记录(count > 0)
				assert.Greater(t, mf.GetMetric()[0].GetHistogram().GetSampleCount(), uint64(0))
			case "test_http_request_size_bytes":
				foundReqSize = true
			case "test_http_response_size_bytes":
				foundRespSize = true
			}
		}

		assert.True(t, foundTotal, "http_requests_total not found")
		assert.True(t, foundDuration, "http_request_duration_seconds not found")
		assert.True(t, foundReqSize, "http_request_size_bytes not found")
		assert.True(t, foundRespSize, "http_response_size_bytes not found")
	})

	t.Run("middleware records different HTTP methods", func(t *testing.T) {
		m := createTestMetrics()
		registry := prometheus.NewRegistry()
		registry.MustRegister(m.HTTPRequestsTotal)
		registry.MustRegister(m.HTTPRequestDuration)
		registry.MustRegister(m.HTTPRequestSize)
		registry.MustRegister(m.HTTPResponseSize)

		router := setupTestRouter()
		router.Use(PrometheusMiddleware(m))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "GET"})
		})
		router.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"message": "POST"})
		})
		router.PUT("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "PUT"})
		})
		router.DELETE("/test", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})

		// 发送不同方法的请求
		methods := []struct {
			method string
			status int
		}{
			{"GET", http.StatusOK},
			{"POST", http.StatusCreated},
			{"PUT", http.StatusOK},
			{"DELETE", http.StatusNoContent},
		}

		for _, m := range methods {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(m.method, "/test", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, m.status, w.Code)
		}

		// 验证指标
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			if mf.GetName() == "test_http_requests_total" {
				// 应该有4个不同的标签组合
				assert.Len(t, mf.GetMetric(), 4)
			}
		}
	})

	t.Run("middleware records different status codes", func(t *testing.T) {
		m := createTestMetrics()
		registry := prometheus.NewRegistry()
		registry.MustRegister(m.HTTPRequestsTotal)
		registry.MustRegister(m.HTTPRequestDuration)
		registry.MustRegister(m.HTTPRequestSize)
		registry.MustRegister(m.HTTPResponseSize)

		router := setupTestRouter()
		router.Use(PrometheusMiddleware(m))
		router.GET("/success", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		router.GET("/created", func(c *gin.Context) {
			c.Status(http.StatusCreated)
		})
		router.GET("/notfound", func(c *gin.Context) {
			c.Status(http.StatusNotFound)
		})
		router.GET("/error", func(c *gin.Context) {
			c.Status(http.StatusInternalServerError)
		})

		// 发送请求
		paths := []struct {
			path   string
			status int
		}{
			{"/success", http.StatusOK},
			{"/created", http.StatusCreated},
			{"/notfound", http.StatusNotFound},
			{"/error", http.StatusInternalServerError},
		}

		for _, p := range paths {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", p.path, nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, p.status, w.Code)
		}

		// 验证指标包含不同的状态码
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			if mf.GetName() == "test_http_requests_total" {
				// 应该有4个不同的路径
				assert.Len(t, mf.GetMetric(), 4)

				// 验证每个指标都有正确的标签
				for _, metric := range mf.GetMetric() {
					labels := metric.GetLabel()
					var hasMethod, hasPath, hasStatus bool
					for _, label := range labels {
						switch label.GetName() {
						case "method":
							hasMethod = true
							assert.Equal(t, "GET", label.GetValue())
						case "path":
							hasPath = true
						case "status":
							hasStatus = true
							// 验证状态码是有效的
							assert.Contains(t, []string{"200", "201", "404", "500"}, label.GetValue())
						}
					}
					assert.True(t, hasMethod, "method label not found")
					assert.True(t, hasPath, "path label not found")
					assert.True(t, hasStatus, "status label not found")
				}
			}
		}
	})

	t.Run("middleware handles 404 paths", func(t *testing.T) {
		m := createTestMetrics()
		registry := prometheus.NewRegistry()
		registry.MustRegister(m.HTTPRequestsTotal)
		registry.MustRegister(m.HTTPRequestDuration)
		registry.MustRegister(m.HTTPRequestSize)
		registry.MustRegister(m.HTTPResponseSize)

		router := setupTestRouter()
		router.Use(PrometheusMiddleware(m))
		// 不注册任何路由,所有请求都是404

		// 发送请求到不存在的路径
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		// 验证指标使用了请求的 URL 路径
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			if mf.GetName() == "test_http_requests_total" {
				require.Len(t, mf.GetMetric(), 1)

				// 验证路径标签
				labels := mf.GetMetric()[0].GetLabel()
				var pathLabel string
				for _, label := range labels {
					if label.GetName() == "path" {
						pathLabel = label.GetValue()
					}
				}
				assert.Equal(t, "/nonexistent", pathLabel)
			}
		}
	})

	t.Run("middleware measures request duration", func(t *testing.T) {
		m := createTestMetrics()
		registry := prometheus.NewRegistry()
		registry.MustRegister(m.HTTPRequestsTotal)
		registry.MustRegister(m.HTTPRequestDuration)
		registry.MustRegister(m.HTTPRequestSize)
		registry.MustRegister(m.HTTPResponseSize)

		router := setupTestRouter()
		router.Use(PrometheusMiddleware(m))
		router.GET("/test", func(c *gin.Context) {
			// 模拟一些处理时间
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 发送请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// 验证耗时被记录
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			if mf.GetName() == "test_http_request_duration_seconds" {
				require.Len(t, mf.GetMetric(), 1)

				histogram := mf.GetMetric()[0].GetHistogram()
				assert.Equal(t, uint64(1), histogram.GetSampleCount())
				assert.Greater(t, histogram.GetSampleSum(), 0.0)
			}
		}
	})

	t.Run("middleware counts multiple requests to same endpoint", func(t *testing.T) {
		m := createTestMetrics()
		registry := prometheus.NewRegistry()
		registry.MustRegister(m.HTTPRequestsTotal)
		registry.MustRegister(m.HTTPRequestDuration)
		registry.MustRegister(m.HTTPRequestSize)
		registry.MustRegister(m.HTTPResponseSize)

		router := setupTestRouter()
		router.Use(PrometheusMiddleware(m))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 发送多次请求
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// 验证计数器
		metrics, err := registry.Gather()
		require.NoError(t, err)

		for _, mf := range metrics {
			if mf.GetName() == "test_http_requests_total" {
				require.Len(t, mf.GetMetric(), 1)
				assert.Equal(t, 5.0, mf.GetMetric()[0].GetCounter().GetValue())
			}
			if mf.GetName() == "test_http_request_duration_seconds" {
				require.Len(t, mf.GetMetric(), 1)
				assert.Equal(t, uint64(5), mf.GetMetric()[0].GetHistogram().GetSampleCount())
			}
		}
	})
}

func TestComputeRequestSize(t *testing.T) {
	t.Run("compute request size", func(t *testing.T) {
		// 测试 computeRequestSize 函数
		req, _ := http.NewRequest("GET", "/test", nil)
		size := computeRequestSize(req)

		// 当前实现返回 0
		assert.Equal(t, int64(0), size)
	})

	t.Run("compute request size with any type", func(t *testing.T) {
		// 测试函数接受任意类型
		size1 := computeRequestSize("string")
		size2 := computeRequestSize(123)
		size3 := computeRequestSize(nil)

		// 所有应该返回 0
		assert.Equal(t, int64(0), size1)
		assert.Equal(t, int64(0), size2)
		assert.Equal(t, int64(0), size3)
	})
}

func TestPrometheusMiddleware_WithDifferentPaths(t *testing.T) {
	m := createTestMetrics()
	registry := prometheus.NewRegistry()
	registry.MustRegister(m.HTTPRequestsTotal)
	registry.MustRegister(m.HTTPRequestDuration)
	registry.MustRegister(m.HTTPRequestSize)
	registry.MustRegister(m.HTTPResponseSize)

	router := setupTestRouter()
	router.Use(PrometheusMiddleware(m))

	// 注册多个路径
	router.GET("/api/agents", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"agents": []string{}})
	})
	router.GET("/api/configurations", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"configs": []string{}})
	})
	router.POST("/api/agents", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": "123"})
	})

	// 发送请求到不同路径
	requests := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/api/agents", http.StatusOK},
		{"GET", "/api/configurations", http.StatusOK},
		{"POST", "/api/agents", http.StatusCreated},
		{"GET", "/api/agents", http.StatusOK}, // 重复请求
	}

	for _, req := range requests {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(req.method, req.path, nil)
		router.ServeHTTP(w, r)
		assert.Equal(t, req.status, w.Code)
	}

	// 验证指标
	metrics, err := registry.Gather()
	require.NoError(t, err)

	for _, mf := range metrics {
		if mf.GetName() == "test_http_requests_total" {
			// 应该有3个不同的标签组合 (GET /api/agents, GET /api/configurations, POST /api/agents)
			assert.Len(t, mf.GetMetric(), 3)

			// GET /api/agents 应该有2次请求
			for _, metric := range mf.GetMetric() {
				labels := metric.GetLabel()
				var isGetAgents bool
				for _, label := range labels {
					if label.GetName() == "method" && label.GetValue() == "GET" {
						for _, l2 := range labels {
							if l2.GetName() == "path" && l2.GetValue() == "/api/agents" {
								isGetAgents = true
							}
						}
					}
				}
				if isGetAgents {
					assert.Equal(t, 2.0, metric.GetCounter().GetValue())
				}
			}
		}
	}
}
