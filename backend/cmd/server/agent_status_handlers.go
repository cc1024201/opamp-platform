package main

import (
	"net/http"
	"strconv"

	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/cc1024201/opamp-platform/internal/store/postgres"
	"github.com/gin-gonic/gin"
)

// getAgentConnectionHistoryHandler 获取 Agent 连接历史
// @Summary      获取 Agent 连接历史
// @Description  获取指定 Agent 的连接历史记录
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
// @Router       /agents/{id}/connection-history [get]
func getAgentConnectionHistoryHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		histories, total, err := store.ListConnectionHistoryByAgent(c.Request.Context(), agentID, limit, offset)
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

// getAgentActiveConnectionHandler 获取 Agent 当前活跃连接
// @Summary      获取 Agent 当前活跃连接
// @Description  获取指定 Agent 当前的活跃连接信息
// @Tags         agents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Agent ID"
// @Success      200 {object} model.AgentConnectionHistory
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /agents/{id}/active-connection [get]
func getAgentActiveConnectionHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("id")

		history, err := store.GetActiveConnectionHistory(c.Request.Context(), agentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if history == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active connection found"})
			return
		}

		c.JSON(http.StatusOK, history)
	}
}

// listOnlineAgentsHandler 列出所有在线的 Agent
// @Summary      列出在线 Agent
// @Description  获取所有状态为在线的 Agent 列表
// @Tags         agents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /agents/online [get]
func listOnlineAgentsHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		agents, err := store.ListOnlineAgents(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"agents": agents,
			"total":  len(agents),
		})
	}
}

// listOfflineAgentsHandler 列出所有离线的 Agent
// @Summary      列出离线 Agent
// @Description  获取所有状态为离线的 Agent 列表,支持分页
// @Tags         agents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "每页数量" default(20)
// @Param        offset query int false "偏移量" default(0)
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /agents/offline [get]
func listOfflineAgentsHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		agents, total, err := store.ListOfflineAgents(c.Request.Context(), limit, offset)
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

// getAgentStatusSummaryHandler 获取 Agent 状态统计
// @Summary      获取 Agent 状态统计
// @Description  获取各状态的 Agent 数量统计
// @Tags         agents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /agents/status/summary [get]
func getAgentStatusSummaryHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 获取在线 Agent
		onlineAgents, err := store.ListOnlineAgents(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 获取离线 Agent 总数
		_, offlineTotal, err := store.ListOfflineAgents(ctx, 0, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 获取所有 Agent 总数
		agents, total, err := store.ListAgents(ctx, 0, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 按状态分组统计
		statusCounts := make(map[model.AgentStatus]int)
		for _, agent := range agents {
			statusCounts[agent.Status]++
		}

		c.JSON(http.StatusOK, gin.H{
			"total":         total,
			"online":        len(onlineAgents),
			"offline":       offlineTotal,
			"status_counts": statusCounts,
		})
	}
}
