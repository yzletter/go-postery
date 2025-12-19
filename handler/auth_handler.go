package handler

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type AuthHandler struct {
	authSvc    service.AuthService
	sessionSvc service.SessionService
}

func NewAuthHandler(authSvc service.AuthService, sessionSvc service.SessionService) *AuthHandler {
	return &AuthHandler{
		authSvc:    authSvc,
		sessionSvc: sessionSvc,
	}
}

// Register 用户注册 Handler
func (hdl *AuthHandler) Register(ctx *gin.Context) {
	// 参数校验
	var registerReq user.RegisterRequest
	err := ctx.ShouldBindJSON(&registerReq)
	if err != nil {
		// 参数绑定失败
		slog.Error("Register Param Bind Failed", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 注册用户
	userBriefDTO, err := hdl.authSvc.Register(ctx, registerReq.Name, registerReq.Email, registerReq.PassWord)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 注册私信功能
	err = hdl.sessionSvc.Register(ctx, userBriefDTO.ID)
	if err != nil {
		response.Error(ctx, errno.ErrServerInternal)
		return
	}

	// 根据 UserID 签发双 Token
	accessToken, refreshToken, err := hdl.authSvc.IssueTokens(ctx, userBriefDTO.ID, 0, ctx.Request.UserAgent())
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	ctx.Header("Authorization", "Bearer "+accessToken)
	ctx.SetCookie(conf.RefreshTokenInCookie, refreshToken, conf.RefreshTokenMaxAgeSecs, "/", "localhost", false, true)

	// 返回成功响应
	response.Success(ctx, "注册成功", userBriefDTO)
}

// Login 登录 Handler
func (hdl *AuthHandler) Login(ctx *gin.Context) {
	// 参数校验
	var loginReq user.LoginRequest
	err := ctx.ShouldBindJSON(&loginReq)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 进行登录
	userBriefDTO, err := hdl.authSvc.Login(ctx, loginReq.Name, loginReq.PassWord)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 根据 UserID 签发双 Token
	accessToken, refreshToken, err := hdl.authSvc.IssueTokens(ctx, userBriefDTO.ID, 0, ctx.Request.UserAgent())
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	ctx.Header("Authorization", "Bearer "+accessToken)
	ctx.SetCookie(conf.RefreshTokenInCookie, refreshToken, conf.RefreshTokenMaxAgeSecs, "/", "localhost", false, true)

	// 返回成功响应
	response.Success(ctx, "登录成功", userBriefDTO)
	return
}

// Logout 登出 Handler
func (hdl *AuthHandler) Logout(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	_, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		slog.Error("Get Uid From CTX Failed", "error", err)
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 从 Header 中获取 AccessToken, 从 Cookie 中获取 RefreshToken
	accessToken := ExtractToken(ctx)
	refreshToken := utils.GetValueFromCookie(ctx, conf.RefreshTokenInCookie)

	// 服务端清理双 Token
	if err := hdl.authSvc.ClearTokens(ctx, accessToken, refreshToken); err != nil {
		response.Error(ctx, err)
		return
	}

	// 将双 Token 置空
	ctx.Header("Authorization", "")
	ctx.SetCookie(conf.RefreshTokenInCookie, "", -1, "/", "localhost", false, true)

	response.Success(ctx, "登出成功", nil)
}

// ExtractToken 从上下文取出 tokenString
func ExtractToken(ctx *gin.Context) string {
	headerString := ctx.GetHeader("Authorization")
	headerStringSeg := strings.SplitN(headerString, " ", 2)
	if len(headerStringSeg) != 2 {
		return ""
	}
	return headerStringSeg[1]
}
