package handler

import (
	"log/slog"
	"net/http"
	"strconv"
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
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
		return
	}
	user := database.GetUserByName(loginRequest.Name)
	if user == nil {
		// 根据 name 未找到 user
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
		return
	}
	if user.PassWord != loginRequest.PassWord {
		// 密码不正确
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
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
		ctx.String(http.StatusInternalServerError, "jwt 生成失败")
	} else {
		// 生成成功, 放入 Cookies
		ctx.SetCookie(JWT_COOKIE_NAME, jwtToken, 86400*7, "/", "localhost", false, true)
	}

	// 默认情况下也返回200
	ctx.String(http.StatusOK, "登录成功")
}

// LogoutHandlerFunc 用户登出 Handler
func LogoutHandlerFunc(ctx *gin.Context) {
	// 设置 Cookie
	ctx.SetCookie(JWT_COOKIE_NAME, "", -1, "/", "localhost", false, true)
}

// ModifyPassHandlerFunc 修改密码 Handler
func ModifyPassHandlerFunc(ctx *gin.Context) {
	var modifyPassRequest model.ModifyPasswordRequest
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&modifyPassRequest)
	if err != nil {
		// 参数绑定失败
		ctx.String(http.StatusBadRequest, "密码输入错误")
		return
	}

	uid, ok := ctx.Value(UID_IN_CTX).(int)
	if !ok {
		ctx.String(http.StatusForbidden, "请先登录") // 没有登录
		return
	}

	err = database.UpdatePassword(uid, modifyPassRequest.OldPass, modifyPassRequest.NewPass)
	if err != nil {
		// 密码更改失败
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	// 默认情况下也返回200
	ctx.String(http.StatusOK, "密码修改成功")
}

// RegisterHandlerFunc 用户注册 Handler
func RegisterHandlerFunc(ctx *gin.Context) {
	var registerRequest model.RegisterRequest
	err := ctx.ShouldBind(&registerRequest)
	if err != nil {
		// 参数绑定失败
		ctx.String(http.StatusBadRequest, "用户名或密码错误")
		return
	}

	_, err = database.RegisterUser(registerRequest.Name, registerRequest.PassWord)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
}

// GetUidFromCookie 从 Cookie 中获取 uid (JWT 引入后废弃)
func GetUidFromCookie(ctx *gin.Context) int {
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name == "uid" {
			uid, err := strconv.Atoi(cookie.Value)
			if err == nil {
				return uid
			}
		}
	}
	return 0
}
