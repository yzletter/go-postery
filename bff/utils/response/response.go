package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/errno"
)

// Response 统一定义后端返回数据格式
type Response struct {
	Code int         `json:"code"`           // 业务状态码，0 表示成功，非 0 表示失败
	Msg  string      `json:"msg"`            // 提示信息
	Data interface{} `json:"data,omitempty"` // 具体数据，失败时可以为空
}

// 统一的业务码
const (
	CodeSuccess      = 0     // 成功
	CodeBadRequest   = 40001 // 参数错误
	CodeUnauthorized = 40003 // 未登录/无权限
	CodeServerError  = 50001 // 服务端错误
	// 其他业务错误码
)

// Success 成功
func Success(ctx *gin.Context, msg string, data interface{}) {
	if msg == "" {
		msg = "success"
	}
	ctx.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	})
}

// Error 失败
func Error(ctx *gin.Context, err error) {
	var e *errno.Error
	if errors.As(err, &e) && e != nil {
		failWithHTTP(ctx, e.HTTPStatus, e.Code, e.Msg)
		return
	}

	// 兜底：非 errno.Error 的未知错误
	ctx.JSON(http.StatusInternalServerError, Response{
		Code: CodeServerError,
		Msg:  "系统繁忙，请稍后重试",
		Data: nil,
	})
}

func failWithHTTP(ctx *gin.Context, httpCode int, code int, msg string) {
	ctx.JSON(httpCode, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}
