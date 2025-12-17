package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

const (
	UserIDInCtx = "user_id" // uid 在上下文中的 name
)

type AuthHandler struct {
	AuthSvc service.AuthService
}

// Register 用户注册 Handler
func (hdl *AuthHandler) Register(ctx *gin.Context) {
	var createUserRequest user.CreateRequest
	err := ctx.ShouldBind(&createUserRequest)
	if err != nil {
		// 参数绑定失败
		slog.Error("Param Bind Failed", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	userBriefDTO, err := hdl.AuthSvc.Register(ctx, createUserRequest.Name, createUserRequest.Email, createUserRequest.PassWord)
	if err != nil {
		// 其他错误均对外暴露为系统繁忙，请稍后重试
		response.ServerError(ctx, "")
		return
	}

	slog.Info("Register User Success", "user", userBriefDTO)

	// 根据 UserID 签发双 Token
	accessToken, refreshToken, err := hdl.AuthSvc.IssueTokens(ctx, userBriefDTO.ID, 0, ctx.Request.UserAgent())
	if err != nil {
		// 双 Token 签发失败
		response.ServerError(ctx, "")
		return
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	ctx.Header(conf.AccessTokenInHeader, accessToken)
	ctx.SetCookie(conf.RefreshTokenInCookie, refreshToken, conf.RefreshTokenCookieMaxAgeSecs, "/", "localhost", false, true)

	// 默认情况下也返回200
	response.Success(ctx, userBriefDTO)
}

// Login 登录 Handler
func (hdl *AuthHandler) Login(ctx *gin.Context) {
	var loginReq = user.LoginRequest{}
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&loginReq)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "用户名或密码错误")
		return
	}

	// 进行登录
	userBriefDTO, err := hdl.AuthSvc.Login(ctx, loginReq.Name, loginReq.PassWord)
	if err != nil {
		// 根据 name 未找到 user或密码不正确
		response.ParamError(ctx, "用户名或密码错误")
		return
	}

	// 根据 UserID 签发双 Token
	accessToken, refreshToken, err := hdl.AuthSvc.IssueTokens(ctx, userBriefDTO.ID, 0, ctx.Request.UserAgent())
	if err != nil {
		// 双 Token 签发失败
		response.ServerError(ctx, "")
		return
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	ctx.Header(conf.AccessTokenInHeader, accessToken)
	ctx.SetCookie(conf.RefreshTokenInCookie, refreshToken, conf.RefreshTokenCookieMaxAgeSecs, "/", "localhost", false, true)

	// 默认情况下也返回200
	response.Success(ctx, userBriefDTO)
	return
}

// Logout 登出 Handler
func (hdl *AuthHandler) Logout(ctx *gin.Context) {
	// 从 Header 中获取 AccessToken, 从 Cookie 中获取 RefreshToken
	accessToken := ctx.GetHeader("x-jwt-token")
	refreshToken := utils.GetValueFromCookie(ctx, conf.RefreshTokenInCookie)

	// 服务端清理双 Token
	if err := hdl.AuthSvc.ClearTokens(ctx, accessToken, refreshToken); err != nil {
		response.Fail(ctx, response.CodeBadRequest, "")
		return
	}

	// 将双 Token 置空
	ctx.Header(conf.AccessTokenInHeader, "")
	ctx.SetCookie(conf.RefreshTokenInCookie, "", -1, "/", "localhost", false, true)

	response.Success(ctx, nil)
}
