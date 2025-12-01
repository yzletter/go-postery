package handler

import (
	"log/slog"

	"github.com/yzletter/go-postery/dto/request"
	"github.com/yzletter/go-postery/service"
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
		response.ParamError(ctx, "用户名或密码错误")
		return
	}

	user := hdl.UserService.GetByName(loginRequest.Name)

	if user == nil {
		// 根据 name 未找到 user
		response.ParamError(ctx, "用户名或密码错误")

		return
	}
	if user.PassWord != loginRequest.PassWord {
		// 密码不正确
		response.ParamError(ctx, "用户名或密码错误")

		return
	}

	// 将 user info 放入 jwt
	userInfo := request.UserInformation{
		Id:   user.Id,
		Name: user.Name,
	}

	slog.Info("登录成功", "user", userInfo)

	// 签发双 Token
	refreshToken, accessToken, err := hdl.AuthService.IssueTokenPairForUser(userInfo)
	if err != nil {
		// Token 签发失败
		slog.Error("Token 签发失败", "error", err)
		response.ServerError(ctx, "")
	}

	// 将双 Token 放进 Cookie
	ctx.SetCookie(service.REFRESH_TOKEN_COOKIE_NAME, refreshToken, 7*86400, "/", "localhost", false, true)
	ctx.SetCookie(service.ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)

	// 默认情况下也返回200
	response.Success(ctx, gin.H{
		"user": gin.H{
			"name": loginRequest.Name,
		},
	})
}

// Logout 用户登出 Handler
func (hdl *UserHandler) Logout(ctx *gin.Context) {
	// todo
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
		response.ParamError(ctx, "")
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, ok := ctx.Value(service.UID_IN_CTX).(int)
	if !ok {
		// 没有登录
		response.Unauthorized(ctx, "请先登录")
		return
	}

	ok, err = hdl.UserService.UpdatePassword(uid, modifyPassRequest.OldPass, modifyPassRequest.NewPass)

	// todo Error
	if ok == false || err != nil {
		// 密码更改失败
		response.ServerError(ctx, "")
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, nil)
}

// Register 用户注册 Handler
func (hdl *UserHandler) Register(ctx *gin.Context) {
	var registerRequest request.RegisterRequest
	err := ctx.ShouldBind(&registerRequest)
	if err != nil {
		// 参数绑定失败
		response.Success(ctx, nil)
		return
	}

	uid, err := hdl.UserService.Register(registerRequest.Name, registerRequest.PassWord)

	if err != nil {
		response.ServerError(ctx, "")
		return
	}

	// 将 user info 放入 jwt
	userInfo := request.UserInformation{
		Id:   uid,
		Name: registerRequest.Name,
	}

	slog.Info("注册成功", "user", userInfo)

	// 签发双 Token
	refreshToken, accessToken, err := hdl.AuthService.IssueTokenPairForUser(userInfo)
	if err != nil {
		// Token 签发失败
		slog.Error("Token 签发失败", "error", err)
		response.ServerError(ctx, "")
	}

	// 将双 Token 放进 Cookie
	ctx.SetCookie(service.REFRESH_TOKEN_COOKIE_NAME, refreshToken, 7*86400, "/", "localhost", false, true)
	ctx.SetCookie(service.ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)

	// 默认情况下也返回200
	response.Success(ctx, gin.H{
		"user": gin.H{
			"name": registerRequest.Name,
		},
	})

}
