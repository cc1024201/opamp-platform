package main

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/cc1024201/opamp-platform/internal/store/postgres"
)

func main() {
	// 加载配置
	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// 初始化日志
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

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
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// 检查管理员是否已存在
	existingAdmin, err := store.GetUserByUsername(ctx, "admin")
	if err != nil {
		log.Fatalf("Failed to check existing admin: %v", err)
	}

	if existingAdmin != nil {
		fmt.Println("✅ Admin user already exists")
		fmt.Printf("Username: %s\n", existingAdmin.Username)
		fmt.Printf("Email: %s\n", existingAdmin.Email)
		return
	}

	// 创建管理员用户
	admin := &model.User{
		Username: "admin",
		Email:    "admin@opamp.local",
		Password: "admin123", // BeforeCreate hook 会自动哈希
		Role:     "admin",
		IsActive: true,
	}

	if err := store.CreateUser(ctx, admin); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Println("✅ Admin user created successfully!")
	fmt.Println("\n=== Default Admin Credentials ===")
	fmt.Println("Username: admin")
	fmt.Println("Password: admin123")
	fmt.Println("Email: admin@opamp.local")
	fmt.Println("\n⚠️  Please change the password after first login!")
}
