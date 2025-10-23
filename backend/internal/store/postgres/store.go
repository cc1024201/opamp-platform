package postgres

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/cc1024201/opamp-platform/internal/model"
)

// Config PostgreSQL 配置
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Store PostgreSQL 存储实现
type Store struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewStore 创建新的 PostgreSQL 存储
func NewStore(config Config, log *zap.Logger) (*Store, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	store := &Store{
		db:     db,
		logger: log,
	}

	// 自动迁移
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Info("PostgreSQL store initialized")
	return store, nil
}

// migrate 执行数据库迁移
func (s *Store) migrate() error {
	return s.db.AutoMigrate(
		&model.Agent{},
		&model.Configuration{},
		&model.Source{},
		&model.Destination{},
		&model.Processor{},
		&model.User{},
		&model.Package{},
		&model.ConfigurationHistory{},
		&model.ConfigurationApplyHistory{},
		&model.AgentConnectionHistory{},
	)
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetAgent 获取 Agent
func (s *Store) GetAgent(ctx context.Context, agentID string) (*model.Agent, error) {
	var agent model.Agent
	result := s.db.WithContext(ctx).Where("id = ?", agentID).First(&agent)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Agent 不存在
		}
		return nil, result.Error
	}
	return &agent, nil
}

// UpsertAgent 创建或更新 Agent
func (s *Store) UpsertAgent(ctx context.Context, agent *model.Agent) error {
	result := s.db.WithContext(ctx).Save(agent)
	return result.Error
}

// ListAgents 列出所有 Agent
func (s *Store) ListAgents(ctx context.Context, limit, offset int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var total int64

	// 计算总数
	if err := s.db.WithContext(ctx).Model(&model.Agent{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	result := s.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("updated_at DESC").
		Find(&agents)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return agents, total, nil
}

// DeleteAgent 删除 Agent
func (s *Store) DeleteAgent(ctx context.Context, agentID string) error {
	result := s.db.WithContext(ctx).Delete(&model.Agent{}, "id = ?", agentID)
	return result.Error
}

// GetConfiguration 获取 Agent 的配置
func (s *Store) GetConfiguration(ctx context.Context, agentID string) (*model.Configuration, error) {
	// 获取 Agent
	agent, err := s.GetAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, nil
	}

	// 如果 Agent 指定了配置名称，直接返回该配置
	if agent.ConfigurationName != "" {
		var config model.Configuration
		result := s.db.WithContext(ctx).Where("name = ?", agent.ConfigurationName).First(&config)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return nil, nil
			}
			return nil, result.Error
		}
		return &config, nil
	}

	// 否则，查找匹配 Agent 标签的配置
	var configs []model.Configuration
	result := s.db.WithContext(ctx).Find(&configs)
	if result.Error != nil {
		return nil, result.Error
	}

	// 查找第一个匹配的配置
	for _, config := range configs {
		if config.MatchesAgent(agent) {
			return &config, nil
		}
	}

	return nil, nil
}

// CreateConfiguration 创建配置
func (s *Store) CreateConfiguration(ctx context.Context, config *model.Configuration) error {
	config.UpdateHash()
	result := s.db.WithContext(ctx).Create(config)
	return result.Error
}

// UpdateConfiguration 更新配置
func (s *Store) UpdateConfiguration(ctx context.Context, config *model.Configuration) error {
	// 在事务中执行
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取当前配置
		var existing model.Configuration
		if err := tx.Where("name = ?", config.Name).First(&existing).Error; err != nil {
			return err
		}

		// 更新配置哈希
		config.UpdateHash()

		// 如果配置内容发生变化,创建历史记录并递增版本号
		if existing.ConfigHash != config.ConfigHash {
			// 保存当前版本到历史记录
			history := &model.ConfigurationHistory{
				ConfigurationName: existing.Name,
				Version:           existing.Version,
				ContentType:       existing.ContentType,
				RawConfig:         existing.RawConfig,
				ConfigHash:        existing.ConfigHash,
				Selector:          existing.Selector,
				Platform:          existing.Platform,
				CreatedAt:         existing.UpdatedAt,
			}
			if err := tx.Create(history).Error; err != nil {
				return err
			}

			// 递增版本号
			config.Version = existing.Version + 1
		} else {
			// 配置内容未变化,保持版本号
			config.Version = existing.Version
		}

		// 保存更新后的配置
		if err := tx.Save(config).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetConfigurationByName 根据名称获取配置
func (s *Store) GetConfigurationByName(ctx context.Context, name string) (*model.Configuration, error) {
	var config model.Configuration
	result := s.db.WithContext(ctx).Where("name = ?", name).First(&config)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &config, nil
}

// ListConfigurations 列出所有配置
func (s *Store) ListConfigurations(ctx context.Context) ([]*model.Configuration, error) {
	var configs []*model.Configuration
	result := s.db.WithContext(ctx).Find(&configs)
	if result.Error != nil {
		return nil, result.Error
	}
	return configs, nil
}

// DeleteConfiguration 删除配置
func (s *Store) DeleteConfiguration(ctx context.Context, name string) error {
	result := s.db.WithContext(ctx).Delete(&model.Configuration{}, "name = ?", name)
	return result.Error
}

// CreateUser 创建用户
func (s *Store) CreateUser(ctx context.Context, user *model.User) error {
	result := s.db.WithContext(ctx).Create(user)
	return result.Error
}

// GetUserByUsername 根据用户名获取用户
func (s *Store) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	result := s.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	result := s.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByID 根据 ID 获取用户
func (s *Store) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	result := s.db.WithContext(ctx).Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

// UpdateUser 更新用户
func (s *Store) UpdateUser(ctx context.Context, user *model.User) error {
	result := s.db.WithContext(ctx).Save(user)
	return result.Error
}

// ListUsers 列出所有用户
func (s *Store) ListUsers(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	result := s.db.WithContext(ctx).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// DeleteUser 删除用户
func (s *Store) DeleteUser(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&model.User{}, id)
	return result.Error
}

// GetDB 获取数据库连接（用于健康检查）
func (s *Store) GetDB() *gorm.DB {
	return s.db
}
