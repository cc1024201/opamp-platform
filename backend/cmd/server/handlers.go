package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/cc1024201/opamp-platform/internal/store/postgres"
)

// Agent handlers

// listAgentsHandler 列出所有 Agent
// @Summary      列出所有 Agent
// @Description  获取所有 Agent 列表，支持分页
// @Tags         agents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "每页数量" default(20)
// @Param        offset query int false "偏移量" default(0)
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /agents [get]
func listAgentsHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取分页参数
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		agents, total, err := store.ListAgents(c.Request.Context(), limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"agents": agents,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})
	}
}

// getAgentHandler 获取单个 Agent
// @Summary      获取 Agent 详情
// @Description  根据 ID 获取单个 Agent 的详细信息
// @Tags         agents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Agent ID"
// @Success      200 {object} model.Agent
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /agents/{id} [get]
func getAgentHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("id")

		agent, err := store.GetAgent(c.Request.Context(), agentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if agent == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
			return
		}

		c.JSON(http.StatusOK, agent)
	}
}

// deleteAgentHandler 删除 Agent
// @Summary      删除 Agent
// @Description  根据 ID 删除指定的 Agent
// @Tags         agents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Agent ID"
// @Success      200 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /agents/{id} [delete]
func deleteAgentHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("id")

		if err := store.DeleteAgent(c.Request.Context(), agentID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "agent deleted"})
	}
}

// Configuration handlers

// listConfigurationsHandler 列出所有配置
// @Summary      列出所有配置
// @Description  获取所有配置列表
// @Tags         configurations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /configurations [get]
func listConfigurationsHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		configs, err := store.ListConfigurations(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"configurations": configs,
			"total":          len(configs),
		})
	}
}

// getConfigurationHandler 获取单个配置
// @Summary      获取配置详情
// @Description  根据名称获取单个配置的详细信息
// @Tags         configurations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name path string true "配置名称"
// @Success      200 {object} model.Configuration
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /configurations/{name} [get]
func getConfigurationHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		config, err := store.GetConfigurationByName(c.Request.Context(), name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if config == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "configuration not found"})
			return
		}

		c.JSON(http.StatusOK, config)
	}
}

// createConfigurationHandler 创建配置
// @Summary      创建新配置
// @Description  创建一个新的配置
// @Tags         configurations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        configuration body model.Configuration true "配置信息"
// @Success      201 {object} model.Configuration
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /configurations [post]
func createConfigurationHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var config model.Configuration
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := store.CreateConfiguration(c.Request.Context(), &config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, config)
	}
}

// updateConfigurationHandler 更新配置
// @Summary      更新配置
// @Description  更新指定名称的配置
// @Tags         configurations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name path string true "配置名称"
// @Param        configuration body model.Configuration true "配置信息"
// @Success      200 {object} model.Configuration
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /configurations/{name} [put]
func updateConfigurationHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		var config model.Configuration
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		config.Name = name

		if err := store.UpdateConfiguration(c.Request.Context(), &config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, config)
	}
}

// deleteConfigurationHandler 删除配置
// @Summary      删除配置
// @Description  根据名称删除指定的配置
// @Tags         configurations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name path string true "配置名称"
// @Success      200 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /configurations/{name} [delete]
func deleteConfigurationHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		if err := store.DeleteConfiguration(c.Request.Context(), name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "configuration deleted"})
	}
}
