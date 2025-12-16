package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

const (
	UserIDInCtx          = "user_id"           // uid 在上下文中的 name
	AccessTokenInCookie  = "jwt-access-token"  // AccessToken 在 cookie 中的 name
	RefreshTokenInCookie = "jwt-refresh-token" // RefreshToken 在 cookie 中的 name
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

		response.ParamError(ctx, "用户名重复")
		return

		// 其他错误均对外暴露为系统繁忙，请稍后重试
		response.ServerError(ctx, "")
		return
	}

	slog.Info("Register User Success", "user", userBriefDTO)

	// 将 user info 放入 jwt 签发双 Token
	refreshToken, accessToken, err := hdl.AuthSvc.IssueTokens(ctx, userBriefDTO.ID, 0)
	if err != nil {
		// 双 Token 签发失败
		slog.Error("Dual Token Issue Failed", "error", err)
		response.ServerError(ctx, "")
		return
	}

	// 将双 Token 放进 Cookie
	ctx.SetCookie(RefreshTokenInCookie, refreshToken, 7*86400, "/", "localhost", false, true)
	ctx.SetCookie(AccessTokenInCookie, accessToken, 0, "/", "localhost", false, true)

	// 默认情况下也返回200
	response.Success(ctx, userBriefDTO)
}

// Login 登录 Handler
func (hdl *AuthHandler) Login(ctx *gin.Context) {
	var loginRequest = user.LoginRequest{}
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&loginRequest)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "用户名或密码错误")
		return
	}

	// 进行登录
	userBriefDTO, err := hdl.AuthSvc.Login(ctx, loginRequest.Name, loginRequest.PassWord)
	if err != nil {
		// 根据 name 未找到 user或密码不正确
		response.ParamError(ctx, "用户名或密码错误")
		return
	}
	slog.Info("登录成功", "userBriefDTO", userBriefDTO.ID)

	// 将 user info 放入 jwt 签发双 Token
	refreshToken, accessToken, err := hdl.AuthSvc.IssueTokens(ctx, userBriefDTO.ID, 0)
	if err != nil {
		// Token 签发失败
		slog.Error("Token 签发失败", "error", err)
		response.ServerError(ctx, "")
	}

	// 将双 Token 放进 Cookie
	ctx.SetCookie(RefreshTokenInCookie, refreshToken, 7*86400, "/", "localhost", false, true)
	ctx.SetCookie(AccessTokenInCookie, accessToken, 0, "/", "localhost", false, true)

	// 默认情况下也返回200
	response.Success(ctx, userBriefDTO)
}

// Logout 登出 Handler
func (hdl *AuthHandler) Logout(ctx *gin.Context) {
	// 设置 Cookie 里的双 Token 都置为 -1
	ctx.SetCookie(RefreshTokenInCookie, "", -1, "/", "localhost", false, true)
	ctx.SetCookie(AccessTokenInCookie, "", -1, "/", "localhost", false, true)

	accessToken := utils.GetValueFromCookie(ctx, AccessTokenInCookie) // 获取 AccessToken
	if err := hdl.AuthSvc.Logout(ctx, accessToken); err != nil {
		response.Fail(ctx, response.CodeBadRequest, "")
		return
	}
	response.Success(ctx, nil)
}
