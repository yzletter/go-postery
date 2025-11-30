package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/dto"
	"github.com/yzletter/go-postery/handler/auth"
	"github.com/yzletter/go-postery/utils"
)

// AuthRequiredMiddleware 强制登录
func AuthRequiredMiddleware(authHandler *auth.AuthHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 尝试通过 AccessToken 认证
		accessToken := utils.GetValueFromCookie(ctx, auth.ACCESS_TOKEN_COOKIE_NAME) // 获取 AccessToken
		userInfo := authHandler.GetUserInfoFromJWT(accessToken)

		// AccessToken 认证直接通过
		if userInfo != nil {
			slog.Info("AuthHandler 认证 AccessToken 成功 ...", "UserInfo", userInfo)

			// 把 userInfo 放入上下文, 以便后续中间件直接使用
			setUserToContext(ctx, userInfo)
			ctx.Next()
		}

		slog.Info("AuthHandler 认证 AccessToken 失败, 尝试认证 RefreshToken ...")

		// AccessToken 认证不通过, 尝试通过 RefreshToken 认证
		refreshToken := utils.GetValueFromCookie(ctx, auth.REFRESH_TOKEN_COOKIE_NAME) // 获取 RefreshToken
		result := authHandler.RedisClient.Get(auth.REFRESH_KEY_PREFIX + refreshToken) // 从 Redis 尝试获取 AccessToken
		if result.Err() != nil {
			// 没拿到 redis 中存的 accessToken, RefreshToken 也认证不通过, 没招了
			slog.Info("AuthHandler 认证 RefreshToken 失败, 需重新登录 ...")
			ctx.Abort() // 当前中间件执行完, 后续中间件不执行
			return
		}

		// 如果 redis 能拿到, 重新放到 Cookie 中
		accessToken = result.Val()
		userInfo = authHandler.GetUserInfoFromJWT(accessToken)
		if userInfo == nil {
			// 虽然拿到了, 但是有问题 (很小概率)
			slog.Error("AuthHandler 从 Redis 中获取到错误的 AccessToken ...", "user", userInfo)
			ctx.Abort() // 当前中间件执行完, 后续中间件不执行
			return
		}

		// 拿到了, 并且也没问题, 放到 Cookie 中
		ctx.SetCookie(auth.ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)
		slog.Info("AuthHandler 认证 RefreshToken 成功 ...", "UserInfo", userInfo)

		// 把 userInfo 放入上下文, 以便后续中间件直接使用
		setUserToContext(ctx, userInfo)
		ctx.Next()
	}
}

// AuthOptionalMiddleware 非强制要求登录
func AuthOptionalMiddleware(authHandler *auth.AuthHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 尝试通过 AccessToken 认证
		accessToken := utils.GetValueFromCookie(ctx, auth.ACCESS_TOKEN_COOKIE_NAME) // 获取 AccessToken
		userInfo := authHandler.GetUserInfoFromJWT(accessToken)

		// AccessToken 认证直接通过
		if userInfo != nil {
			slog.Info("AuthHandler 认证 AccessToken 成功 ...", "UserInfo", userInfo)

			// 把 userInfo 放入上下文, 以便后续中间件直接使用
			setUserToContext(ctx, userInfo)
			ctx.Next()
		}

		slog.Info("AuthHandler 认证 AccessToken 失败, 尝试认证 RefreshToken ...")

		// AccessToken 认证不通过, 尝试通过 RefreshToken 认证
		refreshToken := utils.GetValueFromCookie(ctx, auth.REFRESH_TOKEN_COOKIE_NAME) // 获取 RefreshToken
		result := authHandler.RedisClient.Get(auth.REFRESH_KEY_PREFIX + refreshToken) // 从 Redis 尝试获取 AccessToken
		if result.Err() != nil {
			// 没拿到 redis 中存的 accessToken, RefreshToken 也认证不通过, 没招了
			slog.Info("AuthHandler 认证 RefreshToken 失败, 需重新登录 ...")
			ctx.Next()
			return
		}

		// 如果 redis 能拿到, 重新放到 Cookie 中
		accessToken = result.Val()
		userInfo = authHandler.GetUserInfoFromJWT(accessToken)
		if userInfo == nil {
			// 虽然拿到了, 但是有问题 (很小概率)
			slog.Error("AuthHandler 从 Redis 中获取到错误的 AccessToken ...", "user", userInfo)
			ctx.Next()
			return
		}

		// 拿到了, 并且也没问题, 放到 Cookie 中
		ctx.SetCookie(auth.ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)
		slog.Info("AuthHandler 认证 RefreshToken 成功 ...", "UserInfo", userInfo)

		// 把 userInfo 放入上下文, 以便后续中间件直接使用
		setUserToContext(ctx, userInfo)
		ctx.Next()
	}
}

// 将用户信息放入上下文
func setUserToContext(ctx *gin.Context, userInfo *dto.UserInformation) {
	ctx.Set(auth.UID_IN_CTX, userInfo.Id)
	ctx.Set(auth.UNAME_IN_CTX, userInfo.Name)
	slog.Info("用户信息放入上下文成功 ...", "UserInfo", userInfo)
}
