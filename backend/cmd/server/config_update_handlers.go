package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/cc1024201/opamp-platform/internal/opamp"
	"github.com/cc1024201/opamp-platform/internal/store/postgres"
)

// pushConfigurationHandler 手动推送配置到 Agent
// @Summary      推送配置到 Agent
// @Description  手动触发将配置推送到指定 Agent 或所有匹配的 Agent
// @Tags         configurations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name path string true "配置名称"
// @Param        agent_id query string false "Agent ID (为空则推送到所有匹配的 Agent)"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /configurations/{name}/push [post]
func pushConfigurationHandler(store *postgres.Store, opampServer opamp.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		configName := c.Param("name")
		agentID := c.Query("agent_id")

		// 获取配置
		config, err := store.GetConfigurationByName(c.Request.Context(), configName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if config == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "configuration not found"})
			return
		}

		var affectedAgents []string
		var failedAgents []string

		if agentID != "" {
			// 推送到指定 Agent
			if err := pushConfigToAgent(c.Request.Context(), store, opampServer, agentID, config); err != nil {
				failedAgents = append(failedAgents, agentID)
			} else {
				affectedAgents = append(affectedAgents, agentID)
			}
		} else {
			// 推送到所有匹配的 Agent
			agents, _, err := store.ListAgents(c.Request.Context(), 1000, 0)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			for _, agent := range agents {
				// 检查 Agent 是否匹配配置的选择器
				if !config.MatchesAgent(agent) {
					continue
				}

				// 只推送到已连接的 Agent
				if !opampServer.Connected(agent.ID) {
					continue
				}

				if err := pushConfigToAgent(c.Request.Context(), store, opampServer, agent.ID, config); err != nil {
					failedAgents = append(failedAgents, agent.ID)
				} else {
					affectedAgents = append(affectedAgents, agent.ID)
				}
			}
		}

		// 更新配置的最后应用时间
		now := time.Now()
		config.LastAppliedAt = &now
		if err := store.UpdateConfiguration(c.Request.Context(), config); err != nil {
			// 记录错误但不影响响应
			c.Header("X-Warning", "Failed to update last_applied_at: "+err.Error())
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         "configuration push initiated",
			"affected_agents": affectedAgents,
			"failed_agents":   failedAgents,
			"total":           len(affectedAgents),
			"failed":          len(failedAgents),
		})
	}
}

// pushConfigToAgent 推送配置到单个 Agent
func pushConfigToAgent(ctx context.Context, store *postgres.Store, opampServer opamp.Server, agentID string, config *model.Configuration) error {
	// 创建应用历史记录
	applyHistory := &model.ConfigurationApplyHistory{
		AgentID:           agentID,
		ConfigurationName: config.Name,
		ConfigHash:        config.ConfigHash,
		Status:            model.ApplyStatusApplying,
	}
	if err := store.CreateApplyHistory(ctx, applyHistory); err != nil {
		return err
	}

	// 发送配置到 Agent
	update := &model.AgentUpdate{
		Configuration: config,
	}
	if err := opampServer.SendUpdate(ctx, agentID, update); err != nil {
		// 更新应用历史为失败状态
		applyHistory.Status = model.ApplyStatusFailed
		applyHistory.ErrorMessage = err.Error()
		_ = store.UpdateApplyHistory(ctx, applyHistory)
		return err
	}

	// 注意: 实际的应用成功状态会在 Agent 回复配置状态时更新
	// 这里只标记为 applying 状态

	return nil
}

// listConfigurationHistoryHandler 列出配置的历史版本
// @Summary      列出配置历史版本
// @Description  获取指定配置的所有历史版本
// @Tags         configurations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name path string true "配置名称"
// @Param        limit query int false "每页数量" default(20)
// @Param        offset query int false "偏移量" default(0)
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /configurations/{name}/history [get]
func listConfigurationHistoryHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		configName := c.Param("name")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		histories, total, err := store.ListConfigurationHistory(c.Request.Context(), configName, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"histories": histories,
			"total":     total,
			"limit":     limit,
			"offset":    offset,
		})
	}
}

// getConfigurationHistoryHandler 获取指定版本的配置
// @Summary      获取配置历史版本详情
// @Description  获取指定配置的指定版本详情
// @Tags         configurations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name path string true "配置名称"
// @Param        version path int true "版本号"
// @Success      200 {object} model.ConfigurationHistory
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /configurations/{name}/history/{version} [get]
func getConfigurationHistoryHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		configName := c.Param("name")
		version, err := strconv.Atoi(c.Param("version"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
			return
		}

		history, err := store.GetConfigurationHistory(c.Request.Context(), configName, version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if history == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "history version not found"})
			return
		}

		c.JSON(http.StatusOK, history)
	}
}

// rollbackConfigurationHandler 回滚配置到指定版本
// @Summary      回滚配置
// @Description  将配置回滚到指定的历史版本
// @Tags         configurations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name path string true "配置名称"
// @Param        version path int true "目标版本号"
// @Success      200 {object} model.Configuration
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /configurations/{name}/rollback/{version} [post]
func rollbackConfigurationHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		configName := c.Param("name")
		targetVersion, err := strconv.Atoi(c.Param("version"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
			return
		}

		// 获取目标历史版本
		history, err := store.GetConfigurationHistory(c.Request.Context(), configName, targetVersion)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if history == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "target version not found"})
			return
		}

		// 获取当前配置
		currentConfig, err := store.GetConfigurationByName(c.Request.Context(), configName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if currentConfig == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "configuration not found"})
			return
		}

		// 使用历史版本的内容更新当前配置
		currentConfig.ContentType = history.ContentType
		currentConfig.RawConfig = history.RawConfig
		currentConfig.Selector = history.Selector
		currentConfig.Platform = history.Platform
		// UpdateConfiguration 会自动处理版本号递增和历史记录

		if err := store.UpdateConfiguration(c.Request.Context(), currentConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, currentConfig)
	}
}

// listApplyHistoryHandler 列出配置应用历史
// @Summary      列出配置应用历史
// @Description  获取配置推送到 Agent 的应用历史记录
// @Tags         configurations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        name path string true "配置名称"
// @Param        limit query int false "每页数量" default(20)
// @Param        offset query int false "偏移量" default(0)
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /configurations/{name}/apply-history [get]
func listApplyHistoryHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		configName := c.Param("name")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		histories, total, err := store.ListApplyHistoryByConfig(c.Request.Context(), configName, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"histories": histories,
			"total":     total,
			"limit":     limit,
			"offset":    offset,
		})
	}
}

// getAgentApplyHistoryHandler 获取 Agent 的配置应用历史
// @Summary      获取 Agent 配置应用历史
// @Description  获取指定 Agent 的所有配置应用历史记录
// @Tags         agents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Agent ID"
// @Param        limit query int false "每页数量" default(20)
// @Param        offset query int false "偏移量" default(0)
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /agents/{id}/apply-history [get]
func getAgentApplyHistoryHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		histories, total, err := store.ListApplyHistoryByAgent(c.Request.Context(), agentID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"histories": histories,
			"total":     total,
			"limit":     limit,
			"offset":    offset,
		})
	}
}
