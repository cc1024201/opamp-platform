package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewRateLimiter(t *testing.T) {
	t.Run("create rate limiter", func(t *testing.T) {
		r := rate.Limit(10)
		b := 20

		limiter := NewRateLimiter(r, b)

		assert.NotNil(t, limiter)
		assert.Equal(t, r, limiter.rate)
		assert.Equal(t, b, limiter.burst)
		assert.NotNil(t, limiter.limiters)
		assert.Len(t, limiter.limiters, 0)
	})

	t.Run("different parameters", func(t *testing.T) {
		testCases := []struct {
			name  string
			rate  rate.Limit
			burst int
		}{
			{
				name:  "low rate",
				rate:  1,
				burst: 5,
			},
			{
				name:  "high rate",
				rate:  100,
				burst: 200,
			},
			{
				name:  "very low rate",
				rate:  0.1,
				burst: 1,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				limiter := NewRateLimiter(tc.rate, tc.burst)
				assert.Equal(t, tc.rate, limiter.rate)
				assert.Equal(t, tc.burst, limiter.burst)
			})
		}
	})
}

func TestRateLimiter_getLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 20)

	t.Run("get new limiter for IP", func(t *testing.T) {
		ip := "192.168.1.1"
		limiter := rl.getLimiter(ip)

		assert.NotNil(t, limiter)
		assert.Len(t, rl.limiters, 1)
		assert.Contains(t, rl.limiters, ip)
	})

	t.Run("get existing limiter for IP", func(t *testing.T) {
		ip := "192.168.1.2"

		// 第一次获取
		limiter1 := rl.getLimiter(ip)
		// 第二次获取
		limiter2 := rl.getLimiter(ip)

		// 应该返回同一个 limiter 实例
		assert.Equal(t, limiter1, limiter2)
	})

	t.Run("different IPs get different limiters", func(t *testing.T) {
		// 创建新的 RateLimiter 实例,避免之前测试的影响
		newRL := NewRateLimiter(10, 20)
		ip1 := "192.168.1.10"
		ip2 := "192.168.1.11"

		limiter1 := newRL.getLimiter(ip1)
		limiter2 := newRL.getLimiter(ip2)

		// 使用指针比较,确保是不同的实例
		assert.NotSame(t, limiter1, limiter2, "different IPs should have different limiter instances")
		assert.Len(t, newRL.limiters, 2) // 应该只有 2 个

		// 验证每个 IP 对应的 limiter 确实存在
		assert.Contains(t, newRL.limiters, ip1)
		assert.Contains(t, newRL.limiters, ip2)
	})

	t.Run("concurrent access", func(t *testing.T) {
		rl := NewRateLimiter(10, 20)
		ip := "192.168.1.100"
		concurrency := 100

		var wg sync.WaitGroup
		limiters := make([]*rate.Limiter, concurrency)

		// 并发获取 limiter
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				limiters[index] = rl.getLimiter(ip)
			}(i)
		}

		wg.Wait()

		// 所有获取的 limiter 应该是同一个实例
		for i := 1; i < concurrency; i++ {
			assert.Equal(t, limiters[0], limiters[i])
		}

		// 只应该创建一个 limiter
		assert.Len(t, rl.limiters, 1)
	})
}

func TestRateLimiter_Middleware(t *testing.T) {
	t.Run("allow requests under limit", func(t *testing.T) {
		rl := NewRateLimiter(10, 10) // 10 requests per second, burst of 10
		router := setupTestRouter()

		router.Use(rl.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 发送 10 个请求（在 burst 范围内）
		for i := 0; i < 10; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
		}
	})

	t.Run("block requests over limit", func(t *testing.T) {
		rl := NewRateLimiter(1, 5) // 1 request per second, burst of 5
		router := setupTestRouter()

		router.Use(rl.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		ip := "192.168.1.2:12345"

		// 发送 5 个请求（达到 burst 限制）
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		// 第 6 个请求应该被限流
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "请求过于频繁")
	})

	t.Run("different IPs have independent limits", func(t *testing.T) {
		rl := NewRateLimiter(1, 2) // 1 request per second, burst of 2
		router := setupTestRouter()

		router.Use(rl.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		ip1 := "192.168.1.3:12345"
		ip2 := "192.168.1.4:12345"

		// IP1 消耗配额
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip1

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// IP1 第 3 个请求应该被限流
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip1
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// IP2 应该还有配额
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip2
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("rate limit recovery after waiting", func(t *testing.T) {
		rl := NewRateLimiter(5, 1) // 5 requests per second, burst of 1
		router := setupTestRouter()

		router.Use(rl.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		ip := "192.168.1.5:12345"

		// 第一个请求成功
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 立即发送第二个请求，应该被限流
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// 等待 token 恢复 (1/5 秒 = 200ms)
		time.Sleep(250 * time.Millisecond)

		// 现在应该可以再次请求
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("middleware aborts on rate limit", func(t *testing.T) {
		rl := NewRateLimiter(1, 1)
		router := setupTestRouter()

		handlerCalled := false

		router.Use(rl.Middleware())
		router.GET("/test", func(c *gin.Context) {
			handlerCalled = true
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		ip := "192.168.1.6:12345"

		// 消耗配额
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		handlerCalled = false

		// 被限流的请求不应该调用 handler
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.False(t, handlerCalled, "handler should not be called when rate limited")
	})
}

func TestRateLimiter_ClientIP(t *testing.T) {
	t.Run("extract IP from different sources", func(t *testing.T) {
		rl := NewRateLimiter(10, 10)
		router := setupTestRouter()

		router.Use(rl.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ip": c.ClientIP()})
		})

		testCases := []struct {
			name       string
			remoteAddr string
		}{
			{
				name:       "IPv4 with port",
				remoteAddr: "192.168.1.1:54321",
			},
			{
				name:       "different IPv4",
				remoteAddr: "10.0.0.1:12345",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/test", nil)
				req.RemoteAddr = tc.remoteAddr

				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
			})
		}
	})
}
