package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// AuthorizationHeader HTTP Authorization header
	AuthorizationHeader = "Authorization"
	// AuthorizationTypeBearer Bearer token type
	AuthorizationTypeBearer = "bearer"
	// AuthorizationPayloadKey context key for authorization payload
	AuthorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware 创建认证中间件
func AuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(AuthorizationHeader)

		if len(authorizationHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is not provided"})
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unsupported authorization type"})
			return
		}

		accessToken := fields[1]
		claims, err := jwtManager.VerifyToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// 将 claims 存储到 context 中
		c.Set(AuthorizationPayloadKey, claims)
		c.Next()
	}
}

// GetCurrentUser 从 context 中获取当前用户信息
func GetCurrentUser(c *gin.Context) (*Claims, bool) {
	payload, exists := c.Get(AuthorizationPayloadKey)
	if !exists {
		return nil, false
	}

	claims, ok := payload.(*Claims)
	return claims, ok
}

// RequireRole 检查用户角色
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := GetCurrentUser(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// 检查角色
		hasRole := false
		for _, role := range roles {
			if claims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}
