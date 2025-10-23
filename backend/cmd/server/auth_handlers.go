package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/cc1024201/opamp-platform/internal/auth"
	"github.com/cc1024201/opamp-platform/internal/model"
	"github.com/cc1024201/opamp-platform/internal/store/postgres"
)

// loginHandler 登录处理器
// @Summary      用户登录
// @Description  使用用户名和密码登录系统，返回 JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.LoginRequest true "登录信息"
// @Success      200 {object} model.LoginResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /auth/login [post]
func loginHandler(store *postgres.Store, jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 查找用户
		user, err := store.GetUserByUsername(c.Request.Context(), req.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}

		// 检查用户是否激活
		if !user.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user account is disabled"})
			return
		}

		// 验证密码
		if !user.CheckPassword(req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}

		// 生成 JWT token
		token, err := jwtManager.GenerateToken(user.ID, user.Username, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, model.LoginResponse{
			Token: token,
			User:  user,
		})
	}
}

// registerHandler 注册处理器
// @Summary      用户注册
// @Description  注册新用户账号，返回 JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.RegisterRequest true "注册信息"
// @Success      201 {object} model.LoginResponse
// @Failure      400 {object} map[string]string
// @Failure      409 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /auth/register [post]
func registerHandler(store *postgres.Store, jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 检查用户名是否已存在
		existingUser, err := store.GetUserByUsername(c.Request.Context(), req.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if existingUser != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}

		// 检查邮箱是否已存在
		existingEmail, err := store.GetUserByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if existingEmail != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}

		// 创建新用户
		user := &model.User{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password, // BeforeCreate hook 会自动哈希密码
			Role:     "user",       // 默认角色
			IsActive: true,
		}

		if err := store.CreateUser(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		// 生成 JWT token
		token, err := jwtManager.GenerateToken(user.ID, user.Username, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusCreated, model.LoginResponse{
			Token: token,
			User:  user,
		})
	}
}

// meHandler 获取当前用户信息
// @Summary      获取当前用户信息
// @Description  获取当前登录用户的详细信息
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} model.User
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /me [get]
func meHandler(store *postgres.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := auth.GetCurrentUser(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// 从数据库获取最新用户信息
		user, err := store.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
