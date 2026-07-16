package handler

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/bff/dto/auth"
	userdto "github.com/yzletter/go-postery/backend/bff/dto/user"
	"github.com/yzletter/go-postery/backend/bff/errno"
	"github.com/yzletter/go-postery/backend/conf"
	grpcclient "github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/utils"
	"github.com/yzletter/go-postery/backend/utils/response"
	"google.golang.org/grpc/codes"
)

type AuthHandler struct {
	authSvc grpcclient.AuthClient
	codeSvc grpcclient.CodeClient
	userSvc grpcclient.UserClient
}

func NewAuthHandler(authSvc grpcclient.AuthClient, codeSvc grpcclient.CodeClient, userSvc grpcclient.UserClient) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		codeSvc: codeSvc,
		userSvc: userSvc,
	}
}

func (hdl *AuthHandler) RegisterRouter(engine *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// 身份认证模块
	auth := engine.Group("/auth")
	auth.POST("/sms", hdl.SendSMSCode)                // POST /api/v1/auth/sms				发送短信验证码
	auth.POST("/email", hdl.SendEmailCode)            // POST /api/v1/auth/email				发送邮箱验证码
	auth.POST("/login/password", hdl.LoginByPassword) // POST /api/v1/auth/login/password 	手机号码/邮箱 + 密码登录
	auth.POST("/login/phone", hdl.LoginByPhone)       // POST /api/v1/auth/login/phone 		手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册

	authedAuth := auth.Group("")
	authedAuth.Use(authMiddleware)
	authedAuth.POST("/logout", hdl.Logout)                  // POST /api/v1/auth/logout			退出登录
	authedAuth.GET("/status", hdl.Status)                   // GET /api/v1/auth/status			检查登录状态
	authedAuth.POST("/password/update", hdl.UpdatePassword) // POST /api/v1/auth/password/update	修改密码
	authedAuth.POST("/password/set", hdl.SetPassword)       // POST /api/v1/auth/password/set	初始化密码
	authedAuth.GET("/password/status", hdl.HasPassword)     // GET /api/v1/auth/password/status	查询密码状态
	authedAuth.GET("/auth_identity", hdl.GetAuthIdentity)   // GET /api/v1/auth/auth_identity	获取用户的身份认证
}

// LoginByPassword 手机号码/邮箱 + 密码登录
func (hdl *AuthHandler) LoginByPassword(ctx *gin.Context) {
	// 获取参数并校验
	var req auth.LoginByPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 进行登录
	resp, err := hdl.authSvc.Login(ctx, &auth_grpc.LoginRequest{
		Biz:          auth_grpc.LoginBiz_LOGIN_BIZ_Password,
		Identifier:   req.Identifier,
		Verification: req.Password,
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidCredential,
		}, errno.ErrServerInternal), userdto.BriefDTO{})
		return
	}

	// 根据 UserID 签发双 Token
	tokens, err := hdl.authSvc.IssueTokens(ctx, &auth_grpc.IssueTokenRequest{
		UserID:    resp.UserID,
		Role:      0,
		UserAgent: ctx.Request.UserAgent(),
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), userdto.BriefDTO{})
		return
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	setTokens(ctx, tokens.AccessToken, tokens.RefreshToken)

	// 获取用户
	profile, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: resp.UserID})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
			codes.NotFound:        errno.ErrUserNotFound,
		}, errno.ErrServerInternal), userdto.BriefDTO{})
		return
	}

	// 返回成功响应
	response.Success(ctx, "登录成功", userdto.ToBriefDTO(profile))
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
	resp, err := hdl.authSvc.Login(ctx, &auth_grpc.LoginRequest{
		Biz:          auth_grpc.LoginBiz_LOGIN_BIZ_Phone,
		Identifier:   req.Phone,
		Verification: req.Code,
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrPhoneCodeInvalid,
		}, errno.ErrServerInternal), userdto.BriefDTO{})
		return
	}

	// 根据 UserID 签发双 Token
	tokens, err := hdl.authSvc.IssueTokens(ctx, &auth_grpc.IssueTokenRequest{
		UserID:    resp.UserID,
		Role:      0,
		UserAgent: ctx.Request.UserAgent(),
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), userdto.BriefDTO{})
		return
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	setTokens(ctx, tokens.AccessToken, tokens.RefreshToken)

	// 获取用户
	profile, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: resp.UserID})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
			codes.NotFound:        errno.ErrUserNotFound,
		}, errno.ErrServerInternal), userdto.BriefDTO{})
		return
	}

	// 返回成功响应
	response.Success(ctx, "登录成功", userdto.ToBriefDTO(profile))
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
	if _, err := hdl.codeSvc.Send(ctx, &code_grpc.SendCodeRequest{Biz: code_grpc.CodeBiz_CODE_BIZ_Email, Identifier: req.Email}); err != nil {
		slog.Error("发送邮箱验证码失败", "error", err)
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
			codes.AlreadyExists:   errno.ErrSendToFrequent,
		}, errno.ErrServerInternal), gin.H{})
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
	if _, err := hdl.codeSvc.Send(ctx, &code_grpc.SendCodeRequest{Biz: code_grpc.CodeBiz_CODE_BIZ_SMS, Identifier: req.Phone}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
			codes.AlreadyExists:   errno.ErrSendToFrequent,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "发送短信验证码成功", nil)
}

func (hdl *AuthHandler) UpdatePassword(ctx *gin.Context) {
	// 获取参数并校验
	var req auth.UpdatePassRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	if req.OldPass == req.NewPass {
		response.Error(ctx, errno.ErrSamePassword)
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	if _, err = hdl.authSvc.UpdatePassword(ctx, &auth_grpc.UpdatePasswordRequest{
		UserID: uid, NewPassword: req.NewPass, OldPassword: req.OldPass}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrOldPasswordInvalid,
			codes.NotFound:        errno.ErrNotSetPassword,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, "密码修改成功", nil)
}

// SetPassword 初始化密码
func (hdl *AuthHandler) SetPassword(ctx *gin.Context) {
	// 获取参数并校验
	var req auth.SetPassRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	if _, err := hdl.authSvc.SetPassword(ctx, &auth_grpc.SetPasswordRequest{UserID: uid, Password: req.NewPass, Code: req.Code}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidCode,
			codes.Unauthenticated: errno.ErrPhoneNotBound,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "设置密码成功", nil)
}

// HasPassword 查询密码状态
func (hdl *AuthHandler) HasPassword(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 查询是否有密码
	has, err := hdl.authSvc.HasPassword(ctx, &auth_grpc.UserID{UserID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), auth.PassStatusResponse{})
		return
	}

	response.Success(ctx, "获取密码状态成功", auth.PassStatusResponse{HasPassword: has.Result})
	return
}

// GetAuthIdentity 获取用户身份认证
func (hdl *AuthHandler) GetAuthIdentity(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	authIdentity, err := hdl.authSvc.GetAuthIdentityByUID(ctx, &auth_grpc.UserID{UserID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), auth.AuthIdentityResponse{})
		return
	}

	response.Success(ctx, "获取用户身份认证成功", auth.AuthIdentityResponse{
		Phone: authIdentity.Phone,
		Email: authIdentity.Email,
	})
	return
}

// Logout 退出登录
func (hdl *AuthHandler) Logout(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	_, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		slog.Error("Get Uid From CTX Failed", "error", err)
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 从 Header 中获取 AccessToken, 从 Cookie 中获取 RefreshToken
	accessToken := ExtractToken(ctx)
	refreshToken := utils.GetValueFromCookie(ctx, conf.RefreshTokenInCookie)

	// 服务端清理双 Token
	if _, err := hdl.authSvc.ClearTokens(ctx, &auth_grpc.DualTokens{AccessToken: accessToken, RefreshToken: refreshToken}); err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrLogoutFailed), gin.H{})
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
	if _, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext); err != nil {
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

func mapGRPCErr(err error, mapping map[codes.Code]*errno.Error, fallback *errno.Error) *errno.Error {
	if err == nil {
		return nil
	}

	code := errno.GetGRPCErrCode(err)
	if mapping != nil {
		if mapped, ok := mapping[code]; ok && mapped != nil {
			return mapped
		}
	}

	switch code {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return errno.ErrServiceUnavailable
	}

	if fallback != nil {
		return fallback
	}

	return errno.ErrServerInternal
}
