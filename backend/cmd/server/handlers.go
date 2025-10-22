package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/cc1024201/opamp-platform/internal/store/postgres"
)

// Agent handlers

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
