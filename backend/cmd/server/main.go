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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	"go.uber.org/zap"

	"github.com/cc1024201/opamp-platform/internal/auth"
	"github.com/cc1024201/opamp-platform/internal/metrics"
	"github.com/cc1024201/opamp-platform/internal/opamp"
	"github.com/cc1024201/opamp-platform/internal/packagemgr"
	"github.com/cc1024201/opamp-platform/internal/storage"
	"github.com/cc1024201/opamp-platform/internal/store/postgres"
	_ "github.com/cc1024201/opamp-platform/docs" // Swagger 文档
)

// @title           OpAMP Platform API
// @version         1.1.0
// @description     基于 OpenTelemetry OpAMP 协议的 Agent 管理平台
// @termsOfService  https://github.com/cc1024201/opamp-platform

// @contact.name   API Support
// @contact.url    https://github.com/cc1024201/opamp-platform/issues
// @contact.email  admin@opamp.local

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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

	// 创建 JWT 管理器
	jwtSecretKey := viper.GetString("jwt.secret_key")
	if jwtSecretKey == "" {
		jwtSecretKey = "default-secret-key-change-in-production" // 默认密钥，生产环境必须修改
		logger.Warn("JWT secret key not configured, using default key. Please set jwt.secret_key in config")
	}
	jwtDuration := viper.GetDuration("jwt.duration")
	if jwtDuration == 0 {
		jwtDuration = 24 * time.Hour // 默认 24 小时
	}
	jwtManager := auth.NewJWTManager(jwtSecretKey, jwtDuration)

	// 初始化 MinIO 客户端
	minioConfig := storage.Config{
		Endpoint:  viper.GetString("minio.endpoint"),
		AccessKey: viper.GetString("minio.access_key"),
		SecretKey: viper.GetString("minio.secret_key"),
		Bucket:    viper.GetString("minio.bucket"),
		UseSSL:    viper.GetBool("minio.use_ssl"),
	}
	minioClient, err := storage.NewMinIOClient(minioConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize MinIO client", zap.Error(err))
	}

	// 初始化 Package Manager
	packageManager := packagemgr.NewManager(store, minioClient, logger)

	// 初始化 Metrics
	appMetrics := metrics.NewMetrics("opamp_platform")

	// 创建 HTTP 服务器
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware(logger))
	router.Use(metrics.PrometheusMiddleware(appMetrics))

	// 健康检查端点
	router.GET("/health", healthCheckHandler(store.GetDB()))
	router.GET("/health/ready", readinessHandler(store.GetDB()))
	router.GET("/health/live", livenessHandler())

	// Prometheus metrics 端点
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger API 文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 路由组
	api := router.Group("/api/v1")
	{
		// 公开的认证相关 API（不需要 token）
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", loginHandler(store, jwtManager))
			authGroup.POST("/register", registerHandler(store, jwtManager))
		}

		// 需要认证的 API
		authenticated := api.Group("")
		authenticated.Use(auth.AuthMiddleware(jwtManager))
		{
			// 用户信息
			authenticated.GET("/me", meHandler(store))

			// Agent 相关 API
			agents := authenticated.Group("/agents")
			{
				// 状态统计和分组查询(放在前面,避免被 :id 路由捕获)
				agents.GET("/online", listOnlineAgentsHandler(store))
				agents.GET("/offline", listOfflineAgentsHandler(store))
				agents.GET("/status/summary", getAgentStatusSummaryHandler(store))

				// Agent CRUD
				agents.GET("", listAgentsHandler(store))
				agents.GET("/:id", getAgentHandler(store))
				agents.DELETE("/:id", deleteAgentHandler(store))

				// Agent 状态和历史
				agents.GET("/:id/apply-history", getAgentApplyHistoryHandler(store))
				agents.GET("/:id/connection-history", getAgentConnectionHistoryHandler(store))
				agents.GET("/:id/active-connection", getAgentActiveConnectionHandler(store))
			}

			// Configuration 相关 API
			configs := authenticated.Group("/configurations")
			{
				configs.GET("", listConfigurationsHandler(store))
				configs.GET("/:name", getConfigurationHandler(store))
				configs.POST("", createConfigurationHandler(store))
				configs.PUT("/:name", updateConfigurationHandler(store))
				configs.DELETE("/:name", deleteConfigurationHandler(store))

				// 配置热更新相关
				configs.POST("/:name/push", pushConfigurationHandler(store, opampServer))
				configs.GET("/:name/history", listConfigurationHistoryHandler(store))
				configs.GET("/:name/history/:version", getConfigurationHistoryHandler(store))
				configs.POST("/:name/rollback/:version", rollbackConfigurationHandler(store))
				configs.GET("/:name/apply-history", listApplyHistoryHandler(store))
			}

			// Package 相关 API
			packages := authenticated.Group("/packages")
			{
				packages.GET("", listPackagesHandler(packageManager))
				packages.POST("", uploadPackageHandler(packageManager))
				packages.GET("/:id", getPackageHandler(packageManager))
				packages.GET("/:id/download", downloadPackageHandler(packageManager))
				packages.DELETE("/:id", deletePackageHandler(packageManager))
			}
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
