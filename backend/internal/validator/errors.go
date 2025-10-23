package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse 统一错误响应
type ErrorResponse struct {
	Error   string             `json:"error"`
	Details []ValidationError `json:"details,omitempty"`
}

// FormatValidationErrors 格式化验证错误
func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   fieldError.Field(),
				Message: getErrorMessage(fieldError),
			})
		}
	}

	return errors
}

// getErrorMessage 获取友好的错误消息
func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s 是必填字段", fe.Field())
	case "email":
		return fmt.Sprintf("%s 必须是有效的邮箱地址", fe.Field())
	case "min":
		return fmt.Sprintf("%s 最小长度为 %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s 最大长度为 %s", fe.Field(), fe.Param())
	case "uuid":
		return fmt.Sprintf("%s 必须是有效的 UUID", fe.Field())
	default:
		return fmt.Sprintf("%s 验证失败", fe.Field())
	}
}
