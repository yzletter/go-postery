package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

// Success 成功返回（200 + 业务码 0）
func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "success",
		Data: data,
	})
}

// SuccessMsg 成功但自定义提示信息
func SuccessMsg(ctx *gin.Context, msg string, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	})
}

// Fail 业务失败，但不是服务崩了（也是 200，前端看 Code）
func Fail(ctx *gin.Context, code int, msg string) {
	ctx.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// ParamError 参数错误
func ParamError(ctx *gin.Context, msg string) {
	if msg == "" {
		msg = "参数错误"
	}
	ctx.JSON(http.StatusBadRequest, Response{
		Code: CodeBadRequest,
		Msg:  msg,
		Data: nil,
	})
}

// Unauthorized 未登录/无权限
func Unauthorized(ctx *gin.Context, msg string) {
	if msg == "" {
		msg = "未登录或无权限"
	}
	ctx.JSON(http.StatusUnauthorized, Response{
		Code: CodeUnauthorized,
		Msg:  msg,
		Data: nil,
	})
}

// ServerError 服务端错误
func ServerError(ctx *gin.Context, msg string) {
	if msg == "" {
		msg = "系统繁忙，请稍后重试"
	}
	ctx.JSON(http.StatusInternalServerError, Response{
		Code: CodeServerError,
		Msg:  msg,
		Data: nil,
	})
}
