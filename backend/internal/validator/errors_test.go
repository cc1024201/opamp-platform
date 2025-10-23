package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试用的结构体
type TestStruct struct {
	Username string `validate:"required,min=3,max=20"`
	Email    string `validate:"required,email"`
	Age      int    `validate:"min=18,max=100"`
	UUID     string `validate:"uuid"`
}

func TestFormatValidationErrors(t *testing.T) {
	validate := validator.New()

	t.Run("multiple validation errors", func(t *testing.T) {
		testData := TestStruct{
			Username: "ab",      // too short
			Email:    "invalid", // invalid email
			Age:      15,        // too young
			UUID:     "invalid", // invalid UUID
		}

		err := validate.Struct(testData)
		require.Error(t, err)

		errors := FormatValidationErrors(err)

		// 应该有 4 个验证错误
		assert.Len(t, errors, 4)

		// 检查每个错误都有 Field 和 Message
		for _, e := range errors {
			assert.NotEmpty(t, e.Field)
			assert.NotEmpty(t, e.Message)
		}
	})

	t.Run("required field error", func(t *testing.T) {
		testData := TestStruct{
			Username: "", // required field missing
			Email:    "test@example.com",
			Age:      25,
			UUID:     "550e8400-e29b-41d4-a716-446655440000",
		}

		err := validate.Struct(testData)
		require.Error(t, err)

		errors := FormatValidationErrors(err)

		assert.Greater(t, len(errors), 0)

		// 查找 Username 的错误
		var found bool
		for _, e := range errors {
			if e.Field == "Username" {
				assert.Contains(t, e.Message, "必填字段")
				found = true
				break
			}
		}
		assert.True(t, found, "应该找到 Username 的必填错误")
	})

	t.Run("email validation error", func(t *testing.T) {
		testData := TestStruct{
			Username: "testuser",
			Email:    "invalid-email", // invalid email format
			Age:      25,
			UUID:     "550e8400-e29b-41d4-a716-446655440000",
		}

		err := validate.Struct(testData)
		require.Error(t, err)

		errors := FormatValidationErrors(err)

		// 查找 Email 的错误
		var found bool
		for _, e := range errors {
			if e.Field == "Email" {
				assert.Contains(t, e.Message, "邮箱")
				found = true
				break
			}
		}
		assert.True(t, found, "应该找到 Email 的验证错误")
	})

	t.Run("min validation error", func(t *testing.T) {
		testData := TestStruct{
			Username: "ab", // less than min=3
			Email:    "test@example.com",
			Age:      25,
			UUID:     "550e8400-e29b-41d4-a716-446655440000",
		}

		err := validate.Struct(testData)
		require.Error(t, err)

		errors := FormatValidationErrors(err)

		// 查找 Username 的 min 错误
		var found bool
		for _, e := range errors {
			if e.Field == "Username" {
				assert.Contains(t, e.Message, "最小长度")
				assert.Contains(t, e.Message, "3")
				found = true
				break
			}
		}
		assert.True(t, found, "应该找到 Username 的最小长度错误")
	})

	t.Run("max validation error", func(t *testing.T) {
		testData := TestStruct{
			Username: "validusername",
			Email:    "test@example.com",
			Age:      150, // greater than max=100
			UUID:     "550e8400-e29b-41d4-a716-446655440000",
		}

		err := validate.Struct(testData)
		require.Error(t, err)

		errors := FormatValidationErrors(err)

		// 查找 Age 的 max 错误
		var found bool
		for _, e := range errors {
			if e.Field == "Age" {
				assert.Contains(t, e.Message, "最大长度")
				assert.Contains(t, e.Message, "100")
				found = true
				break
			}
		}
		assert.True(t, found, "应该找到 Age 的最大长度错误")
	})

	t.Run("uuid validation error", func(t *testing.T) {
		testData := TestStruct{
			Username: "validusername",
			Email:    "test@example.com",
			Age:      25,
			UUID:     "not-a-uuid", // invalid UUID
		}

		err := validate.Struct(testData)
		require.Error(t, err)

		errors := FormatValidationErrors(err)

		// 查找 UUID 的错误
		var found bool
		for _, e := range errors {
			if e.Field == "UUID" {
				assert.Contains(t, e.Message, "UUID")
				found = true
				break
			}
		}
		assert.True(t, found, "应该找到 UUID 的验证错误")
	})

	t.Run("no validation errors", func(t *testing.T) {
		testData := TestStruct{
			Username: "validusername",
			Email:    "test@example.com",
			Age:      25,
			UUID:     "550e8400-e29b-41d4-a716-446655440000",
		}

		err := validate.Struct(testData)
		require.NoError(t, err)

		errors := FormatValidationErrors(err)
		assert.Len(t, errors, 0)
	})

	t.Run("non-validation error", func(t *testing.T) {
		// 传入一个非 validator.ValidationErrors 的错误
		someError := assert.AnError

		errors := FormatValidationErrors(someError)
		assert.Len(t, errors, 0)
	})
}

func TestGetErrorMessage(t *testing.T) {
	validate := validator.New()

	t.Run("all error tag types", func(t *testing.T) {
		testData := TestStruct{
			Username: "",        // required
			Email:    "invalid", // email
			Age:      10,        // min
			UUID:     "invalid", // uuid
		}

		err := validate.Struct(testData)
		require.Error(t, err)

		errors := FormatValidationErrors(err)
		require.Greater(t, len(errors), 0)

		// 验证所有错误消息都是中文友好提示
		for _, e := range errors {
			assert.NotContains(t, e.Message, "Error:")
			assert.NotContains(t, e.Message, "Field validation")
		}
	})
}

func TestErrorResponse(t *testing.T) {
	t.Run("ErrorResponse structure", func(t *testing.T) {
		resp := ErrorResponse{
			Error: "验证失败",
			Details: []ValidationError{
				{
					Field:   "Username",
					Message: "Username 是必填字段",
				},
				{
					Field:   "Email",
					Message: "Email 必须是有效的邮箱地址",
				},
			},
		}

		assert.Equal(t, "验证失败", resp.Error)
		assert.Len(t, resp.Details, 2)
		assert.Equal(t, "Username", resp.Details[0].Field)
		assert.Equal(t, "Email", resp.Details[1].Field)
	})

	t.Run("ErrorResponse with no details", func(t *testing.T) {
		resp := ErrorResponse{
			Error:   "通用错误",
			Details: nil,
		}

		assert.Equal(t, "通用错误", resp.Error)
		assert.Nil(t, resp.Details)
	})
}

func TestValidationError(t *testing.T) {
	t.Run("ValidationError structure", func(t *testing.T) {
		err := ValidationError{
			Field:   "TestField",
			Message: "测试错误消息",
		}

		assert.Equal(t, "TestField", err.Field)
		assert.Equal(t, "测试错误消息", err.Message)
	})
}
