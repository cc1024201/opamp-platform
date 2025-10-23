package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cc1024201/opamp-platform/internal/model"
)

func TestListAgentsHandler(t *testing.T) {
	store := setupTestStore(t)

	// 创建测试 agents
	agent1 := &model.Agent{
		ID:       uuid.New().String(),
		Name:     "test-agent-1",
		Status:   model.StatusConnected,
		Version:  "1.0.0",
		Hostname: "host1",
	}
	agent2 := &model.Agent{
		ID:       uuid.New().String(),
		Name:     "test-agent-2",
		Status:   model.StatusDisconnected,
		Version:  "1.0.1",
		Hostname: "host2",
	}

	err := store.UpsertAgent(nil, agent1)
	require.NoError(t, err)
	err = store.UpsertAgent(nil, agent2)
	require.NoError(t, err)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "成功列出所有 agents",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "agents")
				assert.Contains(t, response, "total")
				agents := response["agents"].([]interface{})
				assert.GreaterOrEqual(t, len(agents), 2)
			},
		},
		{
			name:           "带分页参数",
			queryParams:    "?limit=1&offset=0",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, float64(1), response["limit"])
				assert.Equal(t, float64(0), response["offset"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.GET("/agents", listAgentsHandler(store))

			req := httptest.NewRequest(http.MethodGet, "/agents"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}

	// 清理
	store.GetDB().Exec("DELETE FROM agents WHERE name LIKE 'test-agent-%'")
}

func TestGetAgentHandler(t *testing.T) {
	store := setupTestStore(t)

	// 创建测试 agent
	agentID := uuid.New().String()
	agent := &model.Agent{
		ID:       agentID,
		Name:     "test-agent-get",
		Status:   model.StatusConnected,
		Version:  "1.0.0",
		Hostname: "test-host",
	}
	err := store.UpsertAgent(nil, agent)
	require.NoError(t, err)

	tests := []struct {
		name           string
		agentID        string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "成功获取 agent",
			agentID:        agentID,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response model.Agent
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, agentID, response.ID)
				assert.Equal(t, "test-agent-get", response.Name)
			},
		},
		{
			name:           "agent 不存在",
			agentID:        uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.GET("/agents/:id", getAgentHandler(store))

			req := httptest.NewRequest(http.MethodGet, "/agents/"+tt.agentID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}

	// 清理
	store.GetDB().Exec("DELETE FROM agents WHERE name = 'test-agent-get'")
}

func TestDeleteAgentHandler(t *testing.T) {
	store := setupTestStore(t)

	// 创建测试 agent
	agentID := uuid.New().String()
	agent := &model.Agent{
		ID:       agentID,
		Name:     "test-agent-delete",
		Status:   model.StatusConnected,
		Version:  "1.0.0",
		Hostname: "test-host",
	}
	err := store.UpsertAgent(nil, agent)
	require.NoError(t, err)

	tests := []struct {
		name           string
		agentID        string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "成功删除 agent",
			agentID:        agentID,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "deleted")
				// 验证已删除
				deletedAgent, _ := store.GetAgent(nil, agentID)
				assert.Nil(t, deletedAgent)
			},
		},
		{
			name:           "删除不存在的 agent",
			agentID:        uuid.New().String(),
			expectedStatus: http.StatusOK, // DELETE 操作幂等
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "deleted")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.DELETE("/agents/:id", deleteAgentHandler(store))

			req := httptest.NewRequest(http.MethodDelete, "/agents/"+tt.agentID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestListConfigurationsHandler(t *testing.T) {
	store := setupTestStore(t)

	// 创建测试 configurations
	config1 := &model.Configuration{
		Name:        "test-config-1",
		DisplayName: "Test Config 1",
		ContentType: "yaml",
		RawConfig:   "test: config1",
		Selector:    map[string]string{"env": "test"},
	}
	config2 := &model.Configuration{
		Name:        "test-config-2",
		DisplayName: "Test Config 2",
		ContentType: "yaml",
		RawConfig:   "test: config2",
		Selector:    map[string]string{"env": "prod"},
	}

	err := store.CreateConfiguration(nil, config1)
	require.NoError(t, err)
	err = store.CreateConfiguration(nil, config2)
	require.NoError(t, err)

	tests := []struct {
		name           string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "成功列出所有 configurations",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "configurations")
				assert.Contains(t, response, "total")
				configs := response["configurations"].([]interface{})
				assert.GreaterOrEqual(t, len(configs), 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.GET("/configurations", listConfigurationsHandler(store))

			req := httptest.NewRequest(http.MethodGet, "/configurations", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}

	// 清理
	store.GetDB().Exec("DELETE FROM configurations WHERE name LIKE 'test-config-%'")
}

func TestGetConfigurationHandler(t *testing.T) {
	store := setupTestStore(t)

	// 创建测试 configuration
	config := &model.Configuration{
		Name:        "test-config-get",
		DisplayName: "Test Config Get",
		ContentType: "yaml",
		RawConfig:   "test: config",
		Selector:    map[string]string{"env": "test"},
	}
	err := store.CreateConfiguration(nil, config)
	require.NoError(t, err)

	tests := []struct {
		name           string
		configName     string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "成功获取 configuration",
			configName:     "test-config-get",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response model.Configuration
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "test-config-get", response.Name)
				assert.Equal(t, "Test Config Get", response.DisplayName)
			},
		},
		{
			name:           "configuration 不存在",
			configName:     "nonexistent-config",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.GET("/configurations/:name", getConfigurationHandler(store))

			req := httptest.NewRequest(http.MethodGet, "/configurations/"+tt.configName, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}

	// 清理
	store.GetDB().Exec("DELETE FROM configurations WHERE name = 'test-config-get'")
}

func TestCreateConfigurationHandler(t *testing.T) {
	store := setupTestStore(t)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "成功创建 configuration",
			requestBody: model.Configuration{
				Name:        fmt.Sprintf("test-config-create-%d", time.Now().Unix()),
				DisplayName: "Test Config Create",
				ContentType: "yaml",
				RawConfig:   "test: config",
				Selector:    map[string]string{"env": "test"},
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response model.Configuration
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Name)
				assert.NotEmpty(t, response.ConfigHash)
			},
		},
		{
			name:           "无效的请求体",
			requestBody:    "invalid json{",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.POST("/configurations", createConfigurationHandler(store))

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/configurations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}

	// 清理
	store.GetDB().Exec("DELETE FROM configurations WHERE name LIKE 'test-config-create-%'")
}

func TestUpdateConfigurationHandler(t *testing.T) {
	store := setupTestStore(t)

	// 创建测试 configuration
	configName := fmt.Sprintf("test-config-update-%d", time.Now().Unix())
	config := &model.Configuration{
		Name:        configName,
		DisplayName: "Test Config Update",
		ContentType: "yaml",
		RawConfig:   "test: config",
		Selector:    map[string]string{"env": "test"},
	}
	err := store.CreateConfiguration(nil, config)
	require.NoError(t, err)

	tests := []struct {
		name           string
		configName     string
		requestBody    interface{}
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "成功更新 configuration",
			configName: configName,
			requestBody: model.Configuration{
				DisplayName: "Updated Config",
				ContentType: "yaml",
				RawConfig:   "test: updated",
				Selector:    map[string]string{"env": "prod"},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response model.Configuration
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "Updated Config", response.DisplayName)
			},
		},
		{
			name:           "无效的请求体",
			configName:     configName,
			requestBody:    "invalid json{",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.PUT("/configurations/:name", updateConfigurationHandler(store))

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/configurations/"+tt.configName, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}

	// 清理
	store.GetDB().Exec("DELETE FROM configurations WHERE name LIKE 'test-config-update-%'")
}

func TestDeleteConfigurationHandler(t *testing.T) {
	store := setupTestStore(t)

	// 创建测试 configuration
	configName := fmt.Sprintf("test-config-delete-%d", time.Now().Unix())
	config := &model.Configuration{
		Name:        configName,
		DisplayName: "Test Config Delete",
		ContentType: "yaml",
		RawConfig:   "test: config",
		Selector:    map[string]string{"env": "test"},
	}
	err := store.CreateConfiguration(nil, config)
	require.NoError(t, err)

	tests := []struct {
		name           string
		configName     string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "成功删除 configuration",
			configName:     configName,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "deleted")
				// 验证已删除
				deletedConfig, _ := store.GetConfigurationByName(nil, configName)
				assert.Nil(t, deletedConfig)
			},
		},
		{
			name:           "删除不存在的 configuration",
			configName:     "nonexistent-config",
			expectedStatus: http.StatusOK, // DELETE 操作幂等
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "deleted")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.DELETE("/configurations/:name", deleteConfigurationHandler(store))

			req := httptest.NewRequest(http.MethodDelete, "/configurations/"+tt.configName, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
