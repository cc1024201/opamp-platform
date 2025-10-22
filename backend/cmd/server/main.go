package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/cc1024201/opamp-platform/internal/opamp"
	"github.com/cc1024201/opamp-platform/internal/store/postgres"
)

func main() {
	// 初始化日志
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// 加载配置
	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal("Failed to read config", zap.Error(err))
	}

	// 初始化数据库
	dbConfig := postgres.Config{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.dbname"),
		SSLMode:  "disable",
	}

	store, err := postgres.NewStore(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize store", zap.Error(err))
	}
	defer store.Close()

	// 创建 OpAMP 服务器
	opampConfig := opamp.Config{
		Endpoint:  viper.GetString("opamp.endpoint"),
		SecretKey: viper.GetString("opamp.secret_key"),
	}

	opampServer, err := opamp.NewServer(opampConfig, store, logger)
	if err != nil {
		logger.Fatal("Failed to create OpAMP server", zap.Error(err))
	}

	// 启动 OpAMP 服务器
	ctx := context.Background()
	if err := opampServer.Start(ctx); err != nil {
		logger.Fatal("Failed to start OpAMP server", zap.Error(err))
	}

	// 创建 HTTP 服务器
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware(logger))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// API 路由组
	api := router.Group("/api/v1")
	{
		// Agent 相关 API
		agents := api.Group("/agents")
		{
			agents.GET("", listAgentsHandler(store))
			agents.GET("/:id", getAgentHandler(store))
			agents.DELETE("/:id", deleteAgentHandler(store))
		}

		// Configuration 相关 API
		configs := api.Group("/configurations")
		{
			configs.GET("", listConfigurationsHandler(store))
			configs.GET("/:name", getConfigurationHandler(store))
			configs.POST("", createConfigurationHandler(store))
			configs.PUT("/:name", updateConfigurationHandler(store))
			configs.DELETE("/:name", deleteConfigurationHandler(store))
		}
	}

	// OpAMP 端点
	router.Any(opampConfig.Endpoint, gin.WrapF(opampServer.Handler()))

	// 启动 HTTP 服务器
	port := viper.GetInt("server.port")
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		logger.Info("Server starting", zap.Int("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := opampServer.Stop(shutdownCtx); err != nil {
		logger.Error("OpAMP server shutdown error", zap.Error(err))
	}

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped")
}

// CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// 日志中间件
func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}
