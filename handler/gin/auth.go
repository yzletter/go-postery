package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/database/redis"
	"github.com/yzletter/go-postery/handler/model"
	"github.com/yzletter/go-postery/utils"
)

const (
	UID_IN_CTX                = "uid" // uid 在上下文中的 name
	UNAME_IN_CTX              = "uname"
	ACCESS_TOKEN_COOKIE_NAME  = "jwt-access-token"  // AccessToken 在 cookie 中的 name
	REFRESH_TOKEN_COOKIE_NAME = "jwt-refresh-token" // RefreshToken 在 cookie 中的 name
	USERINFO_IN_JWT_PAYLOAD   = "userInfo"
	REFRESH_KEY_PREFIX        = "session_"
)

var (
	JWTConfig = utils.InitViper("./conf", "jwt", utils.YAML)
)

// AuthHandlerFunc 身份认证中间件
func AuthHandlerFunc(ctx *gin.Context) {
	// 尝试通过 AccessToken 认证
	accessToken := getTokenFromCookie(ctx, ACCESS_TOKEN_COOKIE_NAME)
	userInfo := getUserInfoFromJWT(accessToken)
	slog.Info("Auth verify AccessToken ...", "user", userInfo)

	if userInfo != nil {
		// AccessToken 认证直接通过
		slog.Info("Auth verify AccessToken succeed", "user", userInfo)
		// 把 userInfo 放入上下文, 以便后续中间件直接使用
		ctx.Set(UID_IN_CTX, userInfo.Id)
		ctx.Set(UNAME_IN_CTX, userInfo.Name)
		return
	}

	// AccessToken 认证不通过, 尝试通过 RefreshToken  认证
	refreshToken := getTokenFromCookie(ctx, REFRESH_TOKEN_COOKIE_NAME)
	result := redis.GoPosteryRedisClient.Get(REFRESH_KEY_PREFIX + refreshToken)
	if result.Err() != nil { // 没拿到 redis 中存的 accessToken
		// RefreshToken 也认证不通过, 没招了
		slog.Info("Auth verify RefreshToken failed")

		ctx.Redirect(http.StatusTemporaryRedirect, "/login") // 未登录, 进行重定向
		ctx.Abort()                                          // 当前中间件执行完, 后续中间件不执行
		return
	}

	// 如果 redis 能拿到, 重新放到 Cookie 中
	accessToken = result.Val()
	userInfo = getUserInfoFromJWT(accessToken)
	if userInfo == nil {
		// 虽然拿到了, 但是有问题 (很小概率)
		slog.Info("Auth verify redis-AccessToken succeed", "user", userInfo)
		ctx.Redirect(http.StatusTemporaryRedirect, "/login") // 未登录, 进行重定向
		ctx.Abort()                                          // 当前中间件执行完, 后续中间件不执行
		return
	} else {
		ctx.SetCookie(ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)
		slog.Info("Auth", "user", userInfo)
		// 把 userInfo 放入上下文, 以便后续中间件直接使用
		ctx.Set(UID_IN_CTX, userInfo.Id)
		ctx.Set(UNAME_IN_CTX, userInfo.Name)
	}
}

// 从 cookie 中获取值
func getTokenFromCookie(ctx *gin.Context, cookieName string) string {
	cookie, err := ctx.Request.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// 从 JWT Token 中获取 uid
func getUserInfoFromJWT(jwtToken string) *model.UserInformation {
	payload, err := utils.VerifyJWT(jwtToken, JWTConfig.GetString("secret")) // 加密的 key 从配置文件中读取
	if err != nil {
		return nil // jwt 校验失败
	}

	slog.Info("成功从 Token 中校验出 payload", "payload", payload)

	for k, v := range payload.UserDefined {
		if k == USERINFO_IN_JWT_PAYLOAD {
			bs, _ := json.Marshal(v)
			var userInfo model.UserInformation
			_ = json.Unmarshal(bs, &userInfo)
			slog.Info("成功从 Token 中拿出 userinfo", "user", userInfo)
			return &userInfo // Json 反序列化 map[string]any 时，数字会被解析成 float64，而不是 int
		}
	}
	return nil // 未找到 uid
}
