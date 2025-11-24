package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	database "github.com/yzletter/go-postery/database/gorm"
	"github.com/yzletter/go-postery/handler/model"
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

	// 设置 Cookie
	ctx.SetCookie("uid", strconv.Itoa(user.Id), 86400, "/", "localhost", false, true)

	// 默认情况下也返回200
	ctx.String(http.StatusOK, "登录成功")
}

// LogoutHandlerFunc 用户登出 Handler
func LogoutHandlerFunc(ctx *gin.Context) {
	// 设置 Cookie
	ctx.SetCookie("uid", "", -1, "/", "localhost", false, true)
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
	uid := GetUidFromCookie(ctx)
	if uid == 0 {
		// 没有登录
		ctx.String(http.StatusBadRequest, "请先登录")
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
