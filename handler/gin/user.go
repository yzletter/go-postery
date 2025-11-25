package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	database "github.com/yzletter/go-postery/database/gorm"
	"github.com/yzletter/go-postery/handler/model"
	"github.com/yzletter/go-postery/utils"
)

// LoginHandlerFunc 用户登录 Handler
func LoginHandlerFunc(ctx *gin.Context) {
	var loginRequest = model.LoginRequest{}
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

	slog.Info("登录成功", "uid", user.Id)

	// 使用 JWT
	payload := utils.JwtPayload{
		Issue:       "yzletter",
		IssueAt:     time.Now().Unix(),                              // 签发日期为当前时间
		Expiration:  time.Now().Add(86400 * 7 * time.Second).Unix(), // 7 天后过期
		UserDefined: map[string]any{"uid": user.Id},                 // 用户自定义字段
	}

	jwtToken, err := utils.GenJWT(payload, JWTConfig.GetString("secret"))
	if err != nil {
		// jwt 生成失败
		slog.Error("jwt 生成失败", "error", err)
		resp := utils.Resp{
			Code: 1,
			Msg:  "jwt 生成失败",
		}
		ctx.JSON(http.StatusInternalServerError, resp)
	} else {
		// 生成成功, 放入 Cookies
		ctx.SetCookie(JWT_COOKIE_NAME, jwtToken, 86400*7, "/", "localhost", false, true)
	}

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

// LogoutHandlerFunc 用户登出 Handler
func LogoutHandlerFunc(ctx *gin.Context) {
	// 设置 Cookie 里的 JWT 为 -1
	ctx.SetCookie(JWT_COOKIE_NAME, "", -1, "/", "localhost", false, true)
	resp := utils.Resp{
		Code: 0,
		Msg:  "登出成功",
	}
	ctx.JSON(http.StatusOK, resp)
}

// ModifyPassHandlerFunc 修改密码 Handler
func ModifyPassHandlerFunc(ctx *gin.Context) {
	var modifyPassRequest model.ModifyPasswordRequest
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
	uid, ok := ctx.Value(UID_IN_CTX).(int)
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

// RegisterHandlerFunc 用户注册 Handler
func RegisterHandlerFunc(ctx *gin.Context) {
	var registerRequest model.RegisterRequest
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

	_, err = database.RegisterUser(registerRequest.Name, registerRequest.PassWord)
	if err != nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  err.Error(),
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	slog.Info("注册成功", "name", registerRequest.Name)

	// 记录注册用户登录态
	user := database.GetUserByName(registerRequest.Name)
	if user == nil {
		// 根据 name 未找到 user
		resp := utils.Resp{
			Code: 1,
			Msg:  "用户名或密码错误",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	// 使用 JWT
	payload := utils.JwtPayload{
		Issue:       "yzletter",
		IssueAt:     time.Now().Unix(),                              // 签发日期为当前时间
		Expiration:  time.Now().Add(86400 * 7 * time.Second).Unix(), // 7 天后过期
		UserDefined: map[string]any{"uid": user.Id},                 // 用户自定义字段
	}

	jwtToken, err := utils.GenJWT(payload, JWTConfig.GetString("secret"))
	if err != nil {
		// jwt 生成失败
		slog.Error("jwt 生成失败", "error", err)
		resp := utils.Resp{
			Code: 1,
			Msg:  "jwt 生成失败",
		}
		ctx.JSON(http.StatusInternalServerError, resp)
	} else {
		// 生成成功, 放入 Cookies
		ctx.SetCookie(JWT_COOKIE_NAME, jwtToken, 86400*7, "/", "localhost", false, true)
	}

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

// GetUidFromCookie 从 Cookie 中获取 uid (JWT 引入后废弃)
//func GetUidFromCookie(ctx *gin.Context) int {
//	for _, cookie := range ctx.Request.Cookies() {
//		if cookie.Name == "uid" {
//			uid, err := strconv.Atoi(cookie.Value)
//			if err == nil {
//				return uid
//			}
//		}
//	}
//	return 0
//}
