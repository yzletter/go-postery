package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/yzletter/go-postery/dto"
	service "github.com/yzletter/go-postery/service/jwt"
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

// AuthHandler 鉴权中间件的 Handler
type AuthHandler struct {
	JwtService  service.JwtService
	RedisClient redis.Cmdable
}

// NewAuthHandler 构造函数
func NewAuthHandler(jwtService service.JwtService, redisClient redis.Cmdable) *AuthHandler {
	return &AuthHandler{
		JwtService:  jwtService,
		RedisClient: redisClient,
	}
}

// Build 返回 gin.HandlerFunc
func (auth *AuthHandler) Build() gin.HandlerFunc {

	// 定义需要返回的 gin.HandlerFunc
	var AuthHandlerFunc func(ctx *gin.Context)

	// gin.HandlerFunc 的具体实现
	AuthHandlerFunc = func(ctx *gin.Context) {
		// 尝试通过 AccessToken 认证
		accessToken := utils.GetValueFromCookie(ctx, ACCESS_TOKEN_COOKIE_NAME)
		userInfo := auth.GetUserInfoFromJWT(accessToken)
		// todo 日志
		slog.Info("Auth verify AccessToken ...", "user", userInfo)

		if userInfo != nil {
			// AccessToken 认证直接通过
			// todo 日志
			slog.Info("Auth verify AccessToken succeed", "user", userInfo)

			// 把 userInfo 放入上下文, 以便后续中间件直接使用
			ctx.Set(UID_IN_CTX, userInfo.Id)
			ctx.Set(UNAME_IN_CTX, userInfo.Name)
			return
		}

		// AccessToken 认证不通过, 尝试通过 RefreshToken  认证
		refreshToken := utils.GetValueFromCookie(ctx, REFRESH_TOKEN_COOKIE_NAME)
		result := auth.RedisClient.Get(REFRESH_KEY_PREFIX + refreshToken)
		if result.Err() != nil { // 没拿到 redis 中存的 accessToken
			// RefreshToken 也认证不通过, 没招了
			// todo 日志
			slog.Info("Auth verify RefreshToken failed")

			ctx.Redirect(http.StatusTemporaryRedirect, "/login") // 未登录, 进行重定向
			ctx.Abort()                                          // 当前中间件执行完, 后续中间件不执行
			return
		}

		// 如果 redis 能拿到, 重新放到 Cookie 中
		accessToken = result.Val()
		userInfo = auth.GetUserInfoFromJWT(accessToken)
		if userInfo == nil {
			// 虽然拿到了, 但是有问题 (很小概率)
			// todo 日志
			slog.Info("Auth verify redis-AccessToken succeed", "user", userInfo)
			ctx.Redirect(http.StatusTemporaryRedirect, "/login") // 未登录, 进行重定向
			ctx.Abort()                                          // 当前中间件执行完, 后续中间件不执行
			return
		} else {
			ctx.SetCookie(ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)
			// todo 日志
			slog.Info("Auth", "user", userInfo)

			// 把 userInfo 放入上下文, 以便后续中间件直接使用
			ctx.Set(UID_IN_CTX, userInfo.Id)
			ctx.Set(UNAME_IN_CTX, userInfo.Name)
		}
	}

	// 返回 gin.HandlerFunc
	return AuthHandlerFunc
}

// GetUserInfoFromJWT 从 JWT Token 中获取 uid
func (auth *AuthHandler) GetUserInfoFromJWT(jwtToken string) *dto.UserInformation {
	// 解析 JWT Token
	payload, err := auth.JwtService.VerifyToken(jwtToken)
	if err != nil {
		// todo 日志
		return nil // jwt 校验失败
	}
	// todo 日志
	slog.Info("成功从 Token 中校验出 payload", "payload", payload)

	// 获取用户信息
	for k, v := range payload.UserDefined {
		if k == USERINFO_IN_JWT_PAYLOAD {
			bs, _ := json.Marshal(v)
			var userInfo dto.UserInformation
			_ = json.Unmarshal(bs, &userInfo)
			slog.Info("成功从 Token 中拿出 userinfo", "user", userInfo)
			return &userInfo // Json 反序列化 map[string]any 时，数字会被解析成 float64，而不是 int
		}
	}
	return nil // 未找到 uid
}
