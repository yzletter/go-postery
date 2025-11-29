package utils

import "github.com/gin-gonic/gin"

// GetValueFromCookie 从 Cookie 中获取值
func GetValueFromCookie(ctx *gin.Context, cookieName string) string {
	cookie, err := ctx.Request.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
