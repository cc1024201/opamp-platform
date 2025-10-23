package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthMiddleware(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour)

	t.Run("valid token", func(t *testing.T) {
		router := setupTestRouter()

		// 添加认证中间件
		router.Use(AuthMiddleware(manager))

		// 添加一个受保护的路由
		router.GET("/protected", func(c *gin.Context) {
			claims, exists := GetCurrentUser(c)
			require.True(t, exists)
			c.JSON(http.StatusOK, gin.H{
				"user_id":  claims.UserID,
				"username": claims.Username,
				"role":     claims.Role,
			})
		})

		// 生成有效 token
		token, err := manager.GenerateToken(1, "testuser", "admin")
		require.NoError(t, err)

		// 创建请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(AuthMiddleware(manager))
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "authorization header is not provided")
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(AuthMiddleware(manager))
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		testCases := []struct {
			name   string
			header string
		}{
			{
				name:   "only token without Bearer",
				header: "some-token",
			},
			{
				name:   "empty header",
				header: "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/protected", nil)
				if tc.header != "" {
					req.Header.Set("Authorization", tc.header)
				}

				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
			})
		}
	})

	t.Run("unsupported authorization type", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(AuthMiddleware(manager))
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Basic some-token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "unsupported authorization type")
	})

	t.Run("invalid token", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(AuthMiddleware(manager))
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("expired token", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(AuthMiddleware(manager))
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// 生成已过期的 token
		expiredManager := NewJWTManager("test-secret", -time.Hour)
		token, err := expiredManager.GenerateToken(1, "testuser", "user")
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("case insensitive bearer", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(AuthMiddleware(manager))
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		token, err := manager.GenerateToken(1, "testuser", "user")
		require.NoError(t, err)

		testCases := []string{
			"Bearer %s",
			"bearer %s",
			"BEARER %s",
			"BeArEr %s",
		}

		for _, format := range testCases {
			t.Run(format, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/protected", nil)
				req.Header.Set("Authorization", fmt.Sprintf(format, token))

				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
			})
		}
	})
}

func TestGetCurrentUser(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour)

	t.Run("user exists in context", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(AuthMiddleware(manager))
		router.GET("/test", func(c *gin.Context) {
			claims, exists := GetCurrentUser(c)
			assert.True(t, exists)
			assert.NotNil(t, claims)
			assert.Equal(t, uint(1), claims.UserID)
			assert.Equal(t, "testuser", claims.Username)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		token, err := manager.GenerateToken(1, "testuser", "user")
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("user not in context", func(t *testing.T) {
		router := setupTestRouter()
		router.GET("/test", func(c *gin.Context) {
			claims, exists := GetCurrentUser(c)
			assert.False(t, exists)
			assert.Nil(t, claims)
			c.JSON(http.StatusOK, gin.H{"message": "no user"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRequireRole(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour)

	t.Run("user has required role", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(AuthMiddleware(manager))
		router.GET("/admin", RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
		})

		token, err := manager.GenerateToken(1, "admin", "admin")
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("user does not have required role", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(AuthMiddleware(manager))
		router.GET("/admin", RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
		})

		token, err := manager.GenerateToken(1, "user", "user")
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "insufficient permissions")
	})

	t.Run("multiple allowed roles", func(t *testing.T) {
		router := setupTestRouter()
		router.Use(AuthMiddleware(manager))
		router.GET("/resource", RequireRole("admin", "moderator"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "access granted"})
		})

		testCases := []struct {
			name           string
			role           string
			expectedStatus int
		}{
			{
				name:           "admin role allowed",
				role:           "admin",
				expectedStatus: http.StatusOK,
			},
			{
				name:           "moderator role allowed",
				role:           "moderator",
				expectedStatus: http.StatusOK,
			},
			{
				name:           "user role forbidden",
				role:           "user",
				expectedStatus: http.StatusForbidden,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				token, err := manager.GenerateToken(1, "testuser", tc.role)
				require.NoError(t, err)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/resource", nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

				router.ServeHTTP(w, req)

				assert.Equal(t, tc.expectedStatus, w.Code)
			})
		}
	})

	t.Run("user not authenticated", func(t *testing.T) {
		router := setupTestRouter()
		// 不添加 AuthMiddleware，直接使用 RequireRole
		router.GET("/admin", RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "user not authenticated")
	})
}

func TestAuthMiddleware_Integration(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour)

	t.Run("complete auth flow", func(t *testing.T) {
		router := setupTestRouter()

		// 公开路由
		router.GET("/public", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "public"})
		})

		// 受保护路由组
		protected := router.Group("/api")
		protected.Use(AuthMiddleware(manager))
		{
			protected.GET("/profile", func(c *gin.Context) {
				claims, _ := GetCurrentUser(c)
				c.JSON(http.StatusOK, gin.H{
					"user_id":  claims.UserID,
					"username": claims.Username,
				})
			})

			protected.GET("/admin", RequireRole("admin"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin only"})
			})
		}

		// 测试公开路由
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/public", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 测试未授权访问受保护路由
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/profile", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// 测试普通用户访问 profile
		userToken, _ := manager.GenerateToken(1, "user", "user")
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/profile", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 测试普通用户访问 admin 路由
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/admin", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)

		// 测试管理员访问 admin 路由
		adminToken, _ := manager.GenerateToken(2, "admin", "admin")
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/admin", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
