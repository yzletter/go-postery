package auth

import (
	"encoding/json"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/yzletter/go-postery/dto"
	service "github.com/yzletter/go-postery/service/jwt"
	"github.com/yzletter/go-postery/utils"
)

// AuthHandler 鉴权中间件的 Handler
type AuthHandler struct {
	JwtService  *service.JwtService // 依赖 JWT 相关服务
	RedisClient redis.Cmdable       // 依赖 Redis 数据库
}

// NewAuthHandler 构造函数
func NewAuthHandler(jwtService *service.JwtService, redisClient redis.Cmdable) *AuthHandler {
	return &AuthHandler{
		JwtService:  jwtService,
		RedisClient: redisClient,
	}
}

// Build 返回 gin.HandlerFunc
func (auth *AuthHandler) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 尝试通过 AccessToken 认证
		accessToken := utils.GetValueFromCookie(ctx, ACCESS_TOKEN_COOKIE_NAME) // 获取 AccessToken
		userInfo := auth.GetUserInfoFromJWT(accessToken)

		// AccessToken 认证直接通过
		if userInfo != nil {
			slog.Info("AuthHandler 认证 AccessToken 成功 ...", "UserInfo", userInfo)

			// 把 userInfo 放入上下文, 以便后续中间件直接使用
			setUserToContext(ctx, userInfo)
			ctx.Next()
		}

		slog.Info("AuthHandler 认证 AccessToken 失败, 尝试认证 RefreshToken ...")

		// AccessToken 认证不通过, 尝试通过 RefreshToken 认证
		refreshToken := utils.GetValueFromCookie(ctx, REFRESH_TOKEN_COOKIE_NAME) // 获取 RefreshToken
		result := auth.RedisClient.Get(REFRESH_KEY_PREFIX + refreshToken)        // 从 Redis 尝试获取 AccessToken
		if result.Err() != nil {
			// 没拿到 redis 中存的 accessToken, RefreshToken 也认证不通过, 没招了
			slog.Info("AuthHandler 认证 RefreshToken 失败, 需重新登录 ...")
			ctx.Abort() // 当前中间件执行完, 后续中间件不执行
			return
		}

		// 如果 redis 能拿到, 重新放到 Cookie 中
		accessToken = result.Val()
		userInfo = auth.GetUserInfoFromJWT(accessToken)
		if userInfo == nil {
			// 虽然拿到了, 但是有问题 (很小概率)
			slog.Error("AuthHandler 从 Redis 中获取到错误的 AccessToken ...", "user", userInfo)
			ctx.Abort() // 当前中间件执行完, 后续中间件不执行
			return
		}

		// 拿到了, 并且也没问题, 放到 Cookie 中
		ctx.SetCookie(ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)
		slog.Info("AuthHandler 认证 RefreshToken 成功 ...", "UserInfo", userInfo)

		// 把 userInfo 放入上下文, 以便后续中间件直接使用
		setUserToContext(ctx, userInfo)
		ctx.Next()
	}
}

// GetUserInfoFromJWT 从 JWT Token 中获取 uid
func (auth *AuthHandler) GetUserInfoFromJWT(jwtToken string) *dto.UserInformation {
	// 校验 JWT Token
	payload, err := auth.JwtService.VerifyToken(jwtToken)
	if err != nil { // JWT Token 校验失败
		slog.Error("AuthHandler 校验 JWT Token 失败 ...", "err", err)
		return nil
	}

	slog.Info("AuthHandler 校验 JWT Token 成功 ...", "payload", payload)

	// 从 payload 的自定义字段中获取用户信息
	for k, v := range payload.UserDefined {
		if k == USERINFO_IN_JWT_PAYLOAD {
			// todo 待优化
			bs, _ := json.Marshal(v)
			var userInfo dto.UserInformation
			_ = json.Unmarshal(bs, &userInfo)
			slog.Info("AuthHandler 获得 UserInfo 成功 ... ", "userInfo", userInfo)
			return &userInfo
		}
	}

	// 未获得用户信息
	slog.Info("AuthHandler 获得 UserInfo 失败 ... ")
	return nil
}

// 将用户信息放入上下文
func setUserToContext(ctx *gin.Context, userInfo *dto.UserInformation) {
	ctx.Set(UID_IN_CTX, userInfo.Id)
	ctx.Set(UNAME_IN_CTX, userInfo.Name)
	slog.Info("用户信息放入上下文成功 ...", "UserInfo", userInfo)
}
