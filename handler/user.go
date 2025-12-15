package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/yzletter/go-postery/dto/request"
	"github.com/yzletter/go-postery/repository/dao"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	AuthService *service.AuthService
	JwtService  *service.JwtService
	UserService *service.UserService
}

// NewUserHandler 构造函数
func NewUserHandler(authService *service.AuthService, jwtService *service.JwtService, userService *service.UserService) *UserHandler {
	return &UserHandler{
		AuthService: authService,
		JwtService:  jwtService,
		UserService: userService,
	}
}

// Login 用户登录 Handler
func (hdl *UserHandler) Login(ctx *gin.Context) {
	var loginRequest = request.LoginRequest{}
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&loginRequest)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "用户名或密码错误")
		return
	}

	// 进行登录
	ok, userBriefDTO := hdl.UserService.Login(ctx, loginRequest.Name, loginRequest.PassWord)
	if !ok {
		// 根据 name 未找到 user或密码不正确
		response.ParamError(ctx, "用户名或密码错误")
		return
	}
	slog.Info("登录成功", "userBriefDTO", userBriefDTO.Id)

	// 将 user info 放入 jwt 签发双 Token
	refreshToken, accessToken, err := hdl.AuthService.IssueTokenForUser(userBriefDTO.Id, userBriefDTO.Name)
	if err != nil {
		// Token 签发失败
		slog.Error("Token 签发失败", "error", err)
		response.ServerError(ctx, "")
	}

	// 将双 Token 放进 Cookie
	ctx.SetCookie(service.REFRESH_TOKEN_COOKIE_NAME, refreshToken, 7*86400, "/", "localhost", false, true)
	ctx.SetCookie(service.ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)

	// 默认情况下也返回200
	response.Success(ctx, userBriefDTO)
}

// Logout 用户登出 Handler
func (hdl *UserHandler) Logout(ctx *gin.Context) {
	// 设置 Cookie 里的双 Token 都置为 -1
	ctx.SetCookie(service.REFRESH_TOKEN_COOKIE_NAME, "", -1, "/", "localhost", false, true)
	ctx.SetCookie(service.ACCESS_TOKEN_COOKIE_NAME, "", -1, "/", "localhost", false, true)
	response.Success(ctx, nil)
}

// ModifyPass 修改密码 Handler
func (hdl *UserHandler) ModifyPass(ctx *gin.Context) {
	var modifyPassRequest request.ModifyPassRequest
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&modifyPassRequest)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := service.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	err = hdl.UserService.UpdatePassword(ctx, uid, modifyPassRequest.OldPass, modifyPassRequest.NewPass)
	if err != nil {
		// 密码更改失败
		response.ServerError(ctx, "")
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, nil)
}

// Register 用户注册 Handler
func (hdl *UserHandler) Register(ctx *gin.Context) {
	var createUserRequest request.CreateUserRequest
	err := ctx.ShouldBind(&createUserRequest)
	if err != nil {
		// 参数绑定失败
		slog.Error("Param Bind Failed", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	userBriefDTO, err := hdl.UserService.Register(ctx, createUserRequest.Name, createUserRequest.PassWord)
	if err != nil {
		if errors.Is(err, service.ErrNameDuplicated) { // 唯一键冲突
			response.ParamError(ctx, "用户名重复")
			return
		}

		// 其他错误均对外暴露为系统繁忙，请稍后重试
		response.ServerError(ctx, "")
		return
	}

	slog.Info("Register User Success", "user", userBriefDTO)

	// 将 user info 放入 jwt 签发双 Token
	refreshToken, accessToken, err := hdl.AuthService.IssueTokenForUser(userBriefDTO.Id, createUserRequest.Name)
	if err != nil {
		// 双 Token 签发失败
		slog.Error("Dual Token Issue Failed", "error", err)
		response.ServerError(ctx, "")
		return
	}

	// 将双 Token 放进 Cookie
	ctx.SetCookie(service.REFRESH_TOKEN_COOKIE_NAME, refreshToken, 7*86400, "/", "localhost", false, true)
	ctx.SetCookie(service.ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)

	// 默认情况下也返回200
	response.Success(ctx, userBriefDTO)
}

func (hdl *UserHandler) Profile(ctx *gin.Context) {
	uid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	ok, userDetailDTO := hdl.UserService.GetDetailById(ctx, int64(uid))
	if !ok {
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, userDetailDTO)
}

func (hdl *UserHandler) ModifyProfile(ctx *gin.Context) {
	var modifyUserProfileRequest request.ModifyProfileRequest
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&modifyUserProfileRequest)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := service.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	err = hdl.UserService.UpdateProfile(ctx, uid, modifyUserProfileRequest)
	if err != nil {
		slog.Error("Error", err)
		if errors.Is(err, dao.ErrRecordNotFound) {
			response.ParamError(ctx, "")
			return
		}

		response.ServerError(ctx, "")
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, nil)
}
