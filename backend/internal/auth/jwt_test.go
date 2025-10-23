package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTManager(t *testing.T) {
	secretKey := "test-secret-key"
	duration := 24 * time.Hour

	manager := NewJWTManager(secretKey, duration)

	require.NotNil(t, manager)
	assert.Equal(t, secretKey, manager.secretKey)
	assert.Equal(t, duration, manager.tokenDuration)
}

func TestJWTManager_GenerateToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour)

	tests := []struct {
		name     string
		userID   uint
		username string
		role     string
	}{
		{
			name:     "admin user",
			userID:   1,
			username: "admin",
			role:     "admin",
		},
		{
			name:     "regular user",
			userID:   2,
			username: "user123",
			role:     "user",
		},
		{
			name:     "user with special characters",
			userID:   3,
			username: "user@example.com",
			role:     "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateToken(tt.userID, tt.username, tt.role)

			require.NoError(t, err)
			require.NotEmpty(t, token)

			// 验证生成的 token 可以被解析
			claims, err := manager.VerifyToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.username, claims.Username)
			assert.Equal(t, tt.role, claims.Role)
		})
	}
}

func TestJWTManager_VerifyToken(t *testing.T) {
	secretKey := "test-secret"
	manager := NewJWTManager(secretKey, 24*time.Hour)

	t.Run("valid token", func(t *testing.T) {
		userID := uint(1)
		username := "testuser"
		role := "admin"

		token, err := manager.GenerateToken(userID, username, role)
		require.NoError(t, err)

		claims, err := manager.VerifyToken(token)
		require.NoError(t, err)
		require.NotNil(t, claims)

		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, role, claims.Role)
		assert.NotNil(t, claims.ExpiresAt)
		assert.NotNil(t, claims.IssuedAt)
		assert.NotNil(t, claims.NotBefore)
	})

	t.Run("expired token", func(t *testing.T) {
		// 创建一个已过期的 token (duration 设置为负数)
		expiredManager := NewJWTManager(secretKey, -time.Hour)
		token, err := expiredManager.GenerateToken(1, "testuser", "user")
		require.NoError(t, err)

		// 使用正常的 manager 验证过期 token
		claims, err := manager.VerifyToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("invalid token format", func(t *testing.T) {
		invalidToken := "invalid.token.format"

		claims, err := manager.VerifyToken(invalidToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("token with wrong secret", func(t *testing.T) {
		// 使用不同的 secret 生成 token
		otherManager := NewJWTManager("different-secret", 24*time.Hour)
		token, err := otherManager.GenerateToken(1, "testuser", "user")
		require.NoError(t, err)

		// 使用原始 manager 验证 (secret 不匹配)
		claims, err := manager.VerifyToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("token with invalid signing method", func(t *testing.T) {
		// 使用不同的签名方法创建 token
		claims := Claims{
			UserID:   1,
			Username: "testuser",
			Role:     "user",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		}

		// 使用 None 签名方法 (不安全,仅用于测试)
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		// 验证应该失败
		verifiedClaims, err := manager.VerifyToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, verifiedClaims)
	})

	t.Run("empty token", func(t *testing.T) {
		claims, err := manager.VerifyToken("")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestJWTManager_TokenExpiry(t *testing.T) {
	t.Run("token expires after duration", func(t *testing.T) {
		// 创建一个 1 秒过期的 token
		manager := NewJWTManager("test-secret", 1*time.Second)
		token, err := manager.GenerateToken(1, "testuser", "user")
		require.NoError(t, err)

		// 立即验证应该成功
		claims, err := manager.VerifyToken(token)
		require.NoError(t, err)
		assert.NotNil(t, claims)

		// 等待 token 过期
		time.Sleep(2 * time.Second)

		// 验证应该失败
		claims, err = manager.VerifyToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestJWTManager_ClaimsFields(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour)
	userID := uint(100)
	username := "testuser"
	role := "admin"

	token, err := manager.GenerateToken(userID, username, role)
	require.NoError(t, err)

	claims, err := manager.VerifyToken(token)
	require.NoError(t, err)

	t.Run("user id is correct", func(t *testing.T) {
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("username is correct", func(t *testing.T) {
		assert.Equal(t, username, claims.Username)
	})

	t.Run("role is correct", func(t *testing.T) {
		assert.Equal(t, role, claims.Role)
	})

	t.Run("issued at is set", func(t *testing.T) {
		assert.NotNil(t, claims.IssuedAt)
		assert.True(t, claims.IssuedAt.Before(time.Now().Add(1*time.Second)))
	})

	t.Run("not before is set", func(t *testing.T) {
		assert.NotNil(t, claims.NotBefore)
		assert.True(t, claims.NotBefore.Before(time.Now().Add(1*time.Second)))
	})

	t.Run("expires at is set", func(t *testing.T) {
		assert.NotNil(t, claims.ExpiresAt)
		assert.True(t, claims.ExpiresAt.After(time.Now()))
	})
}

func TestErrorTypes(t *testing.T) {
	t.Run("ErrInvalidToken is defined", func(t *testing.T) {
		assert.NotNil(t, ErrInvalidToken)
		assert.Equal(t, "invalid token", ErrInvalidToken.Error())
	})

	t.Run("ErrExpiredToken is defined", func(t *testing.T) {
		assert.NotNil(t, ErrExpiredToken)
		assert.Equal(t, "token has expired", ErrExpiredToken.Error())
	})
}
