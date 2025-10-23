package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/cc1024201/opamp-platform/internal/validator"
)

// ErrorHandler 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// 根据错误类型返回不同的响应
			switch err.Type {
			case gin.ErrorTypeBind:
				// 绑定错误（验证失败）
				validationErrors := validator.FormatValidationErrors(err.Err)
				c.JSON(http.StatusBadRequest, validator.ErrorResponse{
					Error:   "请求参数验证失败",
					Details: validationErrors,
				})
			default:
				// 其他错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}

			c.Abort()
			return
		}
	}
}

// RecoverMiddleware 恢复中间件
func RecoverMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务器内部错误",
		})
		c.Abort()
	})
}
