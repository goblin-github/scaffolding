package response

import (
	"errors"
	"net/http"
	"scaffolding/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"scaffolding/internal/errcode"
	validpkg "scaffolding/pkg/validator"
)

type body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Success 返回 code=0 的成功响应。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, body{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// Error 从 err 中提取 errcode.AppError 并格式化输出。
// 如果不是 AppError，兜底为 INTERNAL_ERROR。
func Error(c *gin.Context, err error) {
	if appErr, ok := errors.AsType[*errcode.AppError](err); ok {
		lang := c.GetHeader("Accept-Language")
		msg := config.Translate(lang, appErr.Reason, appErr.Message)

		c.JSON(http.StatusOK, body{
			Code: appErr.Code,
			Msg:  msg,
		})
		return
	}

	c.JSON(http.StatusOK, body{
		Code: errcode.ErrInternal.Code,
		Msg:  errcode.ErrInternal.Message,
	})
}

// ParamError 参数校验错误，接收 Gin binding 的 validator.ValidationErrors。
func ParamError(c *gin.Context, very validator.ValidationErrors) {
	lang := c.GetHeader("Accept-Language")
	msg := validpkg.CollectFieldErrors(lang, very)

	c.JSON(http.StatusOK, body{
		Code: errcode.ErrInvalidParams.Code,
		Msg:  msg,
	})
}
