package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/utils"
)

const (
	UID_IN_JWT      = "uid" // uid 在 jwt 自定义字段的 name
	UID_IN_CTX      = "uid" // uid 在上下文中的 name
	JWT_COOKIE_NAME = "jwt" // jwt 在 cookie 中的 name
)

var (
	JWTConfig = utils.InitViper("./conf", "jwt", utils.YAML)
)

// AuthHandlerFunc 身份认证中间件
func AuthHandlerFunc(ctx *gin.Context) {
	// 获取登录 uid
	jwtToken := getJWTFromCookie(ctx)
	uid := getUidFromJWT(jwtToken)

	// 判断 uid 是否合法
	if uid == 0 {
		ctx.Redirect(http.StatusTemporaryRedirect, "/login") // 未登录, 进行重定向
		ctx.Abort()                                          // 当前中间件执行完, 后续中间件不执行
		return
	}

	// 把 uid 放入上下文, 以便后续中间件直接使用
	ctx.Set(UID_IN_CTX, uid)
}

// 从 cookie 中获取 JWT Token
func getJWTFromCookie(ctx *gin.Context) string {
	jwtToken := ""
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name == JWT_COOKIE_NAME {
			jwtToken = cookie.Value
			break
		}
	}
	return jwtToken
}

// 从 JWT Token 中获取 uid
func getUidFromJWT(jwtToken string) int {
	payload, err := utils.VerifyJWT(jwtToken, JWTConfig.GetString("secret")) // 加密的 key 从配置文件中读取
	if err != nil {
		return 0 // jwt 校验失败
	}
	for k, v := range payload.UserDefined {
		if k == UID_IN_JWT {
			return int(v.(float64)) // Json 反序列化 map[string]any 时，数字会被解析成 float64，而不是 int
		}
	}
	return 0 // 未找到 uid
}
