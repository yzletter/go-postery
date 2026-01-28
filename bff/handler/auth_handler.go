package handler

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	conf2 "github.com/yzletter/go-postery/auth/conf"
	authdto "github.com/yzletter/go-postery/bff/dto/auth"
	userdto "github.com/yzletter/go-postery/bff/dto/user"
	"github.com/yzletter/go-postery/bff/model"
	"github.com/yzletter/go-postery/bff/utils"
	"github.com/yzletter/go-postery/bff/utils/response"
	code_conf "github.com/yzletter/go-postery/code/conf"
	"github.com/yzletter/go-postery/errno"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandler struct {
	authSvc auth_grpc.AuthServiceClient
	codeSvc code_grpc.CodeServiceClient
	userSvc user_grpc.UserServiceClient
}

func NewAuthHandler(authSvc auth_grpc.AuthServiceClient, codeSvc code_grpc.CodeServiceClient, userSvc user_grpc.UserServiceClient) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		codeSvc: codeSvc,
		userSvc: userSvc,
	}
}

// LoginByPassword 手机号码/邮箱 + 密码登录
func (hdl *AuthHandler) LoginByPassword(ctx *gin.Context) {
	// 获取参数并校验
	var req authdto.LoginByPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 进行登录
	userID, err := hdl.authSvc.LoginByPassword(ctx, &auth_grpc.LoginByPasswordRequest{
		Identifier: req.Identifier,
		Password:   req.Password,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 根据 UserID 签发双 Token
	tokens, err := hdl.authSvc.IssueTokens(ctx, &auth_grpc.IssueTokenRequest{
		UserID:    userID.UserID,
		Role:      0,
		UserAgent: ctx.Request.UserAgent(),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	setTokens(ctx, tokens.AccessToken, tokens.RefreshToken)

	// 获取用户
	profile, err := hdl.userSvc.GetProfileById(ctx, &user_grpc.GetProfileByIdRequest{ID: userID.UserID})

	// 返回成功响应
	response.Success(ctx, "登录成功", userdto.ToBriefDTO(profile))
	return
}

// LoginByPhone 手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册
func (hdl *AuthHandler) LoginByPhone(ctx *gin.Context) {
	// 获取参数并校验
	var req authdto.LoginByPhoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 进行登录
	userID, err := hdl.authSvc.LoginByPhone(ctx, &auth_grpc.LoginByPhoneRequest{Phone: req.Phone, Code: req.Code})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 根据 UserID 签发双 Token
	tokens, err := hdl.authSvc.IssueTokens(ctx, &auth_grpc.IssueTokenRequest{
		UserID:    userID.UserID,
		Role:      0,
		UserAgent: ctx.Request.UserAgent(),
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	setTokens(ctx, tokens.AccessToken, tokens.RefreshToken)

	// 获取用户
	profile, err := hdl.userSvc.GetProfileById(ctx, &user_grpc.GetProfileByIdRequest{ID: userID.UserID})

	// 返回成功响应
	response.Success(ctx, "登录成功", userdto.ToBriefDTO(profile))
	return
}

// SendEmailCode 发送邮箱验证码
func (hdl *AuthHandler) SendEmailCode(ctx *gin.Context) {
	// 获取参数并校验
	var req authdto.SendEmailCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 发送邮件
	if _, err := hdl.codeSvc.Send(ctx, &code_grpc.SendCodeRequest{Biz: int64(model.EmailCode), Identifier: req.Email}); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "发送邮箱验证码成功", nil)
}

// SendSMSCode 发送短信验证码
func (hdl *AuthHandler) SendSMSCode(ctx *gin.Context) {
	// 获取参数并校验
	var req authdto.SendSMSCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 发送短信
	if _, err := hdl.codeSvc.Send(ctx, &code_grpc.SendCodeRequest{Biz: int64(model.SMSCode), Identifier: req.Phone}); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "发送短信验证码成功", nil)
}

func (hdl *AuthHandler) UpdatePassword(ctx *gin.Context) {
	// 获取参数并校验
	var req authdto.UpdatePassRequest
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
	uid, err := utils.GetUidFromCTX(ctx, conf2.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	if _, err = hdl.authSvc.UpdatePassword(ctx, &auth_grpc.UpdatePasswordRequest{
		UserID: uid, NewPassword: req.NewPass, OldPassword: req.OldPass}); err != nil {
		response.Error(ctx, err)
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, "密码修改成功", nil)
}

// SetPassword 初始化密码
func (hdl *AuthHandler) SetPassword(ctx *gin.Context) {
	// 获取参数并校验
	var req authdto.SetPassRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf2.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	if _, err := hdl.authSvc.SetPassword(ctx, &auth_grpc.SetPasswordRequest{UserID: uid, Password: req.NewPass, Code: req.Code}); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "设置密码成功", nil)
}

// HasPassword 查询密码状态
func (hdl *AuthHandler) HasPassword(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf2.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 查询是否有密码
	has, err := hdl.authSvc.HasPassword(ctx, &auth_grpc.UserID{UserID: uid})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取密码状态成功", authdto.PassStatusResponse{HasPassword: has.Result})
	return
}

// GetAuthIdentity 获取用户身份认证
func (hdl *AuthHandler) GetAuthIdentity(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf2.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	authIdentity, err := hdl.authSvc.GetAuthIdentityByUID(ctx, &auth_grpc.UserID{UserID: uid})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取用户身份认证成功", authdto.AuthIdentityResponse{
		Phone: authIdentity.Phone,
		Email: authIdentity.Email,
	})
	return
}

// Logout 退出登录
func (hdl *AuthHandler) Logout(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	_, err := utils.GetUidFromCTX(ctx, conf2.UserIDInContext)
	if err != nil {
		slog.Error("Get Uid From CTX Failed", "error", err)
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 从 Header 中获取 AccessToken, 从 Cookie 中获取 RefreshToken
	accessToken := ExtractToken(ctx)
	refreshToken := utils.GetValueFromCookie(ctx, conf2.RefreshTokenInCookie)

	// 服务端清理双 Token
	if _, err := hdl.authSvc.ClearTokens(ctx, &auth_grpc.DualTokens{AccessToken: accessToken, RefreshToken: refreshToken}); err != nil {
		response.Error(ctx, err)
		return
	}

	// 将双 Token 置空
	ctx.Header("Authorization", "")
	ctx.SetCookie(conf2.RefreshTokenInCookie, "", -1, "/", "localhost", false, true)

	response.Success(ctx, "登出成功", nil)
}

// Status 检查登录状态
func (hdl *AuthHandler) Status(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	if _, err := utils.GetUidFromCTX(ctx, conf2.UserIDInContext); err != nil {
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
	if token, err := ctx.Cookie(conf2.AccessTokenInCookie); err == nil {
		return token
	}

	return ""
}

// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
func setTokens(ctx *gin.Context, accessToken, refreshToken string) {
	ctx.Header("Authorization", "Bearer "+accessToken)
	ctx.SetCookie(conf2.RefreshTokenInCookie, refreshToken, conf2.RefreshTokenMaxAgeSecs, "/", "localhost", false, true)
}

func newCodeGrpcConn() *grpc.ClientConn {
	conn, err := grpc.NewClient(
		"localhost:"+code_conf.Port,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 设置传输安全
	)
	if err != nil {
		return nil
	}

	return conn
}
