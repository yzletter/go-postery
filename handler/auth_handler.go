package handler

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/dto/auth"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type AuthHandler struct {
	authSvc service.AuthService
	codeSvc service.CodeService
}

func NewAuthHandler(authSvc service.AuthService, codeSvc service.CodeService) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		codeSvc: codeSvc,
	}
}

// RegisterByPhone 手机号码 + 验证码 + 密码注册
func (hdl *AuthHandler) RegisterByPhone(ctx *gin.Context) {
	// 参数校验
	var req auth.RegisterByPhoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("Register Param Bind Failed", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 注册用户
	userBriefDTO, err := hdl.authSvc.Register(ctx, model.SMSCode, req.Phone, req.Code, req.Password)
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
	setTokens(ctx, accessToken, refreshToken)

	// 返回成功响应
	response.Success(ctx, "注册成功", userBriefDTO)
	return
}

// RegisterByEmail 邮箱 + 验证码 + 密码注册
func (hdl *AuthHandler) RegisterByEmail(ctx *gin.Context) {
	// 参数校验
	var req auth.RegisterByEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("Register Param Bind Failed", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 注册用户
	userBriefDTO, err := hdl.authSvc.Register(ctx, model.EmailCode, req.Email, req.Code, req.Password)
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
	setTokens(ctx, accessToken, refreshToken)

	// 返回成功响应
	response.Success(ctx, "注册成功", userBriefDTO)
	return
}

// LoginByPhoneAndPassword 手机号码 + 密码登录
func (hdl *AuthHandler) LoginByPhoneAndPassword(ctx *gin.Context) {
	// 获取参数并校验
	var req auth.LoginByPhoneAndPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 进行登录
	userBriefDTO, err := hdl.authSvc.LoginByPassword(ctx, model.SMSCode, req.Phone, req.Password)
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
	setTokens(ctx, accessToken, refreshToken)

	// 返回成功响应
	response.Success(ctx, "登录成功", userBriefDTO)
	return
}

// LoginByEmailAndPassword 邮箱 + 密码登录
func (hdl *AuthHandler) LoginByEmailAndPassword(ctx *gin.Context) {
	// 获取参数并校验
	var req auth.LoginByEmailAndPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 进行登录
	userBriefDTO, err := hdl.authSvc.LoginByPassword(ctx, model.EmailCode, req.Email, req.Password)
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
	setTokens(ctx, accessToken, refreshToken)

	// 返回成功响应
	response.Success(ctx, "登录成功", userBriefDTO)
	return
}

// LoginByPhone 手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册
func (hdl *AuthHandler) LoginByPhone(ctx *gin.Context) {
	// 获取参数并校验
	var req auth.LoginByPhoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 进行登录
	userBriefDTO, err := hdl.authSvc.LoginByPhone(ctx, req.Phone, req.Code)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 根据 UID 签发双 Token
	accessToken, refreshToken, err := hdl.authSvc.IssueTokens(ctx, userBriefDTO.ID, 0, ctx.Request.UserAgent())
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	setTokens(ctx, accessToken, refreshToken)

	// 返回成功响应
	response.Success(ctx, "根据手机号登录成功", userBriefDTO)
	return
}

// LoginByEmail 邮箱 + 验证码进行登录
func (hdl *AuthHandler) LoginByEmail(ctx *gin.Context) {
	// 获取参数并校验
	var req auth.LoginByEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 进行登录
	userBriefDTO, err := hdl.authSvc.LoginByEmail(ctx, req.Email, req.Code)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 根据 UID 签发双 Token
	accessToken, refreshToken, err := hdl.authSvc.IssueTokens(ctx, userBriefDTO.ID, 0, ctx.Request.UserAgent())
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	setTokens(ctx, accessToken, refreshToken)

	// 返回成功响应
	response.Success(ctx, "根据邮箱登录成功", userBriefDTO)
	return
}

// SendEmailCode 发送邮箱验证码
func (hdl *AuthHandler) SendEmailCode(ctx *gin.Context) {
	// 获取参数并校验
	var req auth.SendEmailCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 发送邮件
	if err := hdl.codeSvc.SendCode(ctx, model.EmailCode, req.Email); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "发送邮箱验证码成功", nil)
}

// SendSMSCode 发送短信验证码
func (hdl *AuthHandler) SendSMSCode(ctx *gin.Context) {
	// 获取参数并校验
	var req auth.SendSMSCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 发送短信
	if err := hdl.codeSvc.SendCode(ctx, model.SMSCode, req.Phone); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "发送短信验证码成功", nil)
}

// Logout 退出登录
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

// Status 检查登录状态
func (hdl *AuthHandler) Status(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	if _, err := utils.GetUidFromCTX(ctx, UserIDInContext); err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	response.Success(ctx, "登录状态检查成功", nil)
}

// ExtractToken 从上下文取出 tokenString
func ExtractToken(ctx *gin.Context) string {
	//	HTTP 从 Header 中拿 token
	headerString := ctx.GetHeader("Authorization")
	headerStringSeg := strings.SplitN(headerString, " ", 2)

	if len(headerStringSeg) == 2 {
		return headerStringSeg[1]
	}

	// HTTP 从 Cookie 中拿 token
	if token, err := ctx.Cookie(conf.AccessTokenInCookie); err == nil {
		return token
	}

	return ""
}

// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
func setTokens(ctx *gin.Context, accessToken, refreshToken string) {
	ctx.Header("Authorization", "Bearer "+accessToken)
	ctx.SetCookie(conf.RefreshTokenInCookie, refreshToken, conf.RefreshTokenMaxAgeSecs, "/", "localhost", false, true)
}
