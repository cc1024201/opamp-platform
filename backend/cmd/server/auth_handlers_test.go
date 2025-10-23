package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/cc1024201/opamp-platform/internal/auth"
	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/cc1024201/opamp-platform/internal/store/postgres"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func setupTestStore(t *testing.T) *postgres.Store {
	// 创建测试用的 logger
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := postgres.Config{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnvInt("TEST_DB_PORT", 5432),
		User:     getEnv("TEST_DB_USER", "opamp"),
		Password: getEnv("TEST_DB_PASSWORD", "opamp123"),
		DBName:   getEnv("TEST_DB_NAME", "opamp_platform"),
		SSLMode:  "disable",
	}

	store, err := postgres.NewStore(config, logger)
	require.NoError(t, err)

	// 清理测试数据
	t.Cleanup(func() {
		cleanupTestData(store)
		store.Close()
	})

	return store
}

func cleanupTestData(store *postgres.Store) {
	db := store.GetDB()
	db.Exec("DELETE FROM users WHERE username LIKE 'test%'")
}

func getEnv(key, defaultValue string) string {
	// 简化版本，实际可以使用 os.Getenv
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	return defaultValue
}

func TestLoginHandler(t *testing.T) {
	store := setupTestStore(t)
	jwtManager := auth.NewJWTManager("test-secret-key", 24*time.Hour)

	// 创建测试用户
	testUser := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
		IsActive: true,
	}
	err := store.CreateUser(nil, testUser)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "成功登录",
			requestBody: model.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response model.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, "testuser", response.User.Username)
			},
		},
		{
			name: "错误的密码",
			requestBody: model.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "invalid username or password")
			},
		},
		{
			name: "用户不存在",
			requestBody: model.LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "invalid username or password")
			},
		},
		{
			name: "缺少用户名",
			requestBody: map[string]string{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "error")
			},
		},
		{
			name: "缺少密码",
			requestBody: map[string]string{
				"username": "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.POST("/login", loginHandler(store, jwtManager))

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	store := setupTestStore(t)
	jwtManager := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	router := setupTestRouter()
	router.POST("/register", registerHandler(store, jwtManager))

	// 测试成功注册 - 必须先执行以创建基准用户
	t.Run("成功注册", func(t *testing.T) {
		reqBody := model.RegisterRequest{
			Username: "testuser2",
			Email:    "test2@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response model.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.Equal(t, "testuser2", response.User.Username)
		assert.Equal(t, "test2@example.com", response.User.Email)
	})

	// 测试用户名已存在
	t.Run("用户名已存在", func(t *testing.T) {
		reqBody := model.RegisterRequest{
			Username: "testuser2", // 使用上面创建的用户名
			Email:    "another@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "already exists")
	})

	// 测试邮箱已存在
	t.Run("邮箱已存在", func(t *testing.T) {
		reqBody := model.RegisterRequest{
			Username: "testuser3",
			Email:    "test2@example.com", // 使用已存在的邮箱
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "already exists")
	})

	// 测试无效的请求体
	t.Run("无效的请求体", func(t *testing.T) {
		reqBody := map[string]string{
			"username": "test",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

func TestMeHandler(t *testing.T) {
	store := setupTestStore(t)
	jwtManager := auth.NewJWTManager("test-secret-key", 24*time.Hour)

	// 创建测试用户
	testUser := &model.User{
		Username: "testuser4",
		Email:    "test4@example.com",
		Password: "password123",
		Role:     "user",
		IsActive: true,
	}
	err := store.CreateUser(nil, testUser)
	require.NoError(t, err)

	// 生成有效的 token
	validToken, err := jwtManager.GenerateToken(testUser.ID, testUser.Username, testUser.Role)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "成功获取用户信息",
			token:          validToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var user model.User
				err := json.Unmarshal(w.Body.Bytes(), &user)
				require.NoError(t, err)
				assert.Equal(t, "testuser4", user.Username)
				assert.Equal(t, "test4@example.com", user.Email)
			},
		},
		{
			name:           "缺少 token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "authorization header")
			},
		},
		{
			name:           "无效的 token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.GET("/me", auth.AuthMiddleware(jwtManager), meHandler(store))

			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
