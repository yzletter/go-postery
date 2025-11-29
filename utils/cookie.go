package utils

import "github.com/gin-gonic/gin"

// GetValueFromCookie 从 Cookie 中获取值
func GetValueFromCookie(ctx *gin.Context, cookieName string) string {
	// 根据 CookieName 获取 Cookie
	cookie, err := ctx.Request.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
