# OpAMP Platform Backend Makefile

# Database configuration
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= opamp
DB_PASSWORD ?= opamp123
DB_NAME ?= opamp_platform
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Paths
MIGRATIONS_DIR = migrations
BINARY_NAME = opamp-server
BUILD_DIR = bin

# Go commands
GOTEST = go test
GOBUILD = go build
GOMOD = go mod
MIGRATE = ~/go/bin/migrate

.PHONY: help
help: ## 显示帮助信息
	@echo "OpAMP Platform 开发命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## 编译项目
	@echo "编译 $(BINARY_NAME)..."
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: run
run: build ## 运行服务器
	@echo "启动服务器..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: test
test: ## 运行测试
	@echo "运行测试..."
	@$(GOTEST) -v ./internal/...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率..."
	@$(GOTEST) -v -cover -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

.PHONY: swagger
swagger: ## 生成 Swagger 文档
	@echo "生成 Swagger 文档..."
	@~/go/bin/swag init -g cmd/server/main.go --output ./docs
	@echo "Swagger 文档已生成"

.PHONY: deps
deps: ## 安装依赖
	@echo "安装依赖..."
	@$(GOMOD) download
	@$(GOMOD) tidy

.PHONY: clean
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "清理完成"

# Database migration commands

.PHONY: migrate-create
migrate-create: ## 创建新迁移文件 (使用: make migrate-create name=add_users_table)
	@if [ -z "$(name)" ]; then \
		echo "错误: 请指定迁移名称 (例如: make migrate-create name=add_users_table)"; \
		exit 1; \
	fi
	@echo "创建迁移: $(name)..."
	@$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
	@echo "迁移文件已创建"

.PHONY: migrate-up
migrate-up: ## 应用所有待处理的迁移
	@echo "应用数据库迁移..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "迁移完成"

.PHONY: migrate-down
migrate-down: ## 回滚最后一次迁移
	@echo "回滚数据库迁移..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
	@echo "回滚完成"

.PHONY: migrate-force
migrate-force: ## 强制设置迁移版本 (使用: make migrate-force version=1)
	@if [ -z "$(version)" ]; then \
		echo "错误: 请指定版本号 (例如: make migrate-force version=1)"; \
		exit 1; \
	fi
	@echo "强制设置迁移版本为 $(version)..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(version)
	@echo "版本已设置"

.PHONY: migrate-version
migrate-version: ## 显示当前迁移版本
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

.PHONY: migrate-goto
migrate-goto: ## 迁移到指定版本 (使用: make migrate-goto version=1)
	@if [ -z "$(version)" ]; then \
		echo "错误: 请指定版本号 (例如: make migrate-goto version=1)"; \
		exit 1; \
	fi
	@echo "迁移到版本 $(version)..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" goto $(version)
	@echo "迁移完成"

.PHONY: migrate-drop
migrate-drop: ## 删除所有表 (危险!)
	@echo "警告: 这将删除所有表!"
	@read -p "确认删除? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" drop -f; \
		echo "所有表已删除"; \
	else \
		echo "操作已取消"; \
	fi

# Admin commands

.PHONY: create-admin
create-admin: ## 创建管理员用户
	@echo "创建管理员用户..."
	@go run scripts/create_admin.go

# Development commands

.PHONY: dev
dev: ## 开发模式 (自动重新加载)
	@echo "开发模式启动..."
	@air || echo "请安装 air: go install github.com/air-verse/air@latest"

.PHONY: docker-up
docker-up: ## 启动 Docker 服务
	@echo "启动 Docker 服务..."
	@docker-compose up -d
	@echo "Docker 服务已启动"

.PHONY: docker-down
docker-down: ## 停止 Docker 服务
	@echo "停止 Docker 服务..."
	@docker-compose down
	@echo "Docker 服务已停止"

.PHONY: docker-logs
docker-logs: ## 查看 Docker 日志
	@docker-compose logs -f

# All-in-one commands

.PHONY: setup
setup: deps docker-up migrate-up create-admin ## 初始化完整开发环境
	@echo ""
	@echo "✅ 开发环境初始化完成!"
	@echo ""
	@echo "下一步:"
	@echo "  1. 运行 'make run' 启动服务器"
	@echo "  2. 访问 http://localhost:8080/swagger/index.html 查看 API 文档"
	@echo "  3. 使用 admin/admin123 登录"
	@echo ""

.PHONY: reset
reset: docker-down docker-up migrate-drop migrate-up create-admin ## 重置开发环境
	@echo ""
	@echo "✅ 开发环境已重置!"

.DEFAULT_GOAL := help
