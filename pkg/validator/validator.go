package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// TranslateBindingError 把 Gin binding 错误转成对用户友好的多语言提示。
func TranslateBindingError(acceptLanguage, tagName, fieldName, param string) string {
	//lang := config.DetectLang(acceptLanguage)
	return enMessage(tagName, fieldName, param)
}

func enMessage(tag, field, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "email":
		return fmt.Sprintf("%s is not a valid email", field)
	default:
		return fmt.Sprintf("%s validation failed on %s", field, tag)
	}
}

func zhMessage(tag, field, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s不能为空", field)
	case "max":
		return fmt.Sprintf("%s长度不能超过%s", field, param)
	case "min":
		return fmt.Sprintf("%s长度不能少于%s", field, param)
	case "email":
		return fmt.Sprintf("%s不是有效的邮箱", field)
	default:
		return fmt.Sprintf("%s校验失败: %s", field, tag)
	}
}

// CollectFieldErrors 收集所有字段校验错误，用分号拼接。
func CollectFieldErrors(acceptLanguage string, very validator.ValidationErrors) string {
	var msgs []string
	for _, fe := range very {
		msgs = append(msgs, TranslateBindingError(acceptLanguage, fe.Tag(), fe.Field(), fe.Param()))
	}
	if len(msgs) == 0 {
		return "validation failed"
	}
	return strings.Join(msgs, "; ")
}
