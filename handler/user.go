package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/rs/xid"
	"github.com/yzletter/go-postery/dto"
	"github.com/yzletter/go-postery/repository/gorm"
	"github.com/yzletter/go-postery/service"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/utils"
)

type UserHandler struct {
	JwtService  *service.JwtService
	UserService *service.UserService
	RedisClient redis.Cmdable
}

// NewUserHandler 构造函数
func NewUserHandler(redisClient redis.Cmdable, jwtService *service.JwtService, userService *service.UserService) *UserHandler {
	return &UserHandler{
		RedisClient: redisClient,
		JwtService:  jwtService,
		UserService: userService,
	}
}

// Login 用户登录 Handler
func (userHandler *UserHandler) Login(ctx *gin.Context) {
	var loginRequest = dto.LoginRequest{}
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&loginRequest)
	if err != nil {
		// 参数绑定失败
		resp := utils.Resp{
			Code: 1,
			Msg:  "用户名或密码错误",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	user := database.GetUserByName(loginRequest.Name)
	if user == nil {
		// 根据 name 未找到 user
		resp := utils.Resp{
			Code: 1,
			Msg:  "用户名或密码错误",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	if user.PassWord != loginRequest.PassWord {
		// 密码不正确
		resp := utils.Resp{
			Code: 1,
			Msg:  "用户名或密码错误",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 将 user info 放入 jwt
	userInfo := dto.UserInformation{
		Id:   user.Id,
		Name: user.Name,
	}

	slog.Info("登录成功", "user", userInfo)

	// 生成 RefreshToken
	refreshToken := xid.New().String() //	生成一个随机的字符串

	// 生成 AccessToken
	payload := service.JwtPayload{
		Issue:       "yzletter",
		IssueAt:     time.Now().Unix(),                                         // 签发日期为当前时间
		Expiration:  0,                                                         // 永不过期
		UserDefined: map[string]any{service.USERINFO_IN_JWT_PAYLOAD: userInfo}, // 用户自定义字段
	}
	accessToken, err := userHandler.JwtService.GenToken(payload)
	if err != nil {
		// AccessToken 生成失败
		slog.Error("AccessToken 生成失败", "error", err)
		resp := utils.Resp{
			Code: 1,
			Msg:  "AccessToken 生成失败",
		}
		ctx.JSON(http.StatusInternalServerError, resp)
	}

	// 将双 Token 放进 Cookie
	ctx.SetCookie(service.REFRESH_TOKEN_COOKIE_NAME, refreshToken, 7*86400, "/", "localhost", false, true)
	ctx.SetCookie(service.ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)
	// < session_refreshToken, accessToken > 放入 redis
	userHandler.RedisClient.Set(service.REFRESH_KEY_PREFIX+refreshToken, accessToken, 7*86400*time.Second)

	// 默认情况下也返回200
	resp := utils.Resp{
		Code: 0,
		Msg:  "登录成功",
		Data: gin.H{
			"user": gin.H{
				"name": loginRequest.Name,
			},
		},
	}
	ctx.JSON(http.StatusOK, resp)
}

// Logout 用户登出 Handler
func (userHandler *UserHandler) Logout(ctx *gin.Context) {
	// 设置 Cookie 里的双 Token 都置为 -1
	ctx.SetCookie(service.REFRESH_TOKEN_COOKIE_NAME, "", -1, "/", "localhost", false, true)
	ctx.SetCookie(service.ACCESS_TOKEN_COOKIE_NAME, "", -1, "/", "localhost", false, true)

	resp := utils.Resp{
		Code: 0,
		Msg:  "登出成功",
	}
	ctx.JSON(http.StatusOK, resp)
}

// ModifyPass 修改密码 Handler
func (userHandler *UserHandler) ModifyPass(ctx *gin.Context) {
	var modifyPassRequest dto.ModifyPasswordRequest
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&modifyPassRequest)
	if err != nil {
		// 参数绑定失败
		resp := utils.Resp{
			Code: 1,
			Msg:  "参数绑定失败",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, ok := ctx.Value(service.UID_IN_CTX).(int)
	if !ok {
		// 没有登录
		resp := utils.Resp{
			Code: 1,
			Msg:  "请先登录",
		}
		ctx.JSON(http.StatusForbidden, resp)
		return
	}

	err = database.UpdatePassword(uid, modifyPassRequest.OldPass, modifyPassRequest.NewPass)
	if err != nil {
		// 密码更改失败
		resp := utils.Resp{
			Code: 1,
			Msg:  err.Error(),
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 默认情况下也返回200
	resp := utils.Resp{
		Code: 0,
		Msg:  "密码修改成功",
	}
	ctx.JSON(http.StatusOK, resp)
}

// Register 用户注册 Handler
func (userHandler *UserHandler) Register(ctx *gin.Context) {
	var registerRequest dto.RegisterRequest
	err := ctx.ShouldBind(&registerRequest)
	if err != nil {
		// 参数绑定失败
		resp := utils.Resp{
			Code: 1,
			Msg:  "参数绑定失败",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	uid, err := database.RegisterUser(registerRequest.Name, registerRequest.PassWord)
	if err != nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  err.Error(),
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 将 user info 放入 jwt
	userInfo := dto.UserInformation{
		Id:   uid,
		Name: registerRequest.Name,
	}

	slog.Info("注册成功", "user", userInfo)

	// 生成 RefreshToken
	refreshToken := xid.New().String() //	生成一个随机的字符串

	// 生成 AccessToken
	payload := service.JwtPayload{
		Issue:       "yzletter",
		IssueAt:     time.Now().Unix(),                                         // 签发日期为当前时间
		Expiration:  0,                                                         // 永不过期
		UserDefined: map[string]any{service.USERINFO_IN_JWT_PAYLOAD: userInfo}, // 用户自定义字段
	}
	accessToken, err := userHandler.JwtService.GenToken(payload)
	if err != nil {
		// AccessToken 生成失败
		slog.Error("AccessToken 生成失败", "error", err)
		resp := utils.Resp{
			Code: 1,
			Msg:  "AccessToken 生成失败",
		}
		ctx.JSON(http.StatusInternalServerError, resp)
	}

	// 将双 Token 放进 Cookie
	ctx.SetCookie(service.REFRESH_TOKEN_COOKIE_NAME, refreshToken, 7*86400, "/", "localhost", false, true)
	ctx.SetCookie(service.ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)
	// < session_refreshToken, accessToken > 放入 redis
	userHandler.RedisClient.Set(service.REFRESH_KEY_PREFIX+refreshToken, accessToken, 7*86400*time.Second)

	// 默认情况下也返回200
	resp := utils.Resp{
		Code: 0,
		Msg:  "注册成功",
		Data: gin.H{
			"user": gin.H{
				"name": registerRequest.Name,
			},
		},
	}
	ctx.JSON(http.StatusOK, resp)
}
