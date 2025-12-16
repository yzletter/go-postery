package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/handler"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
)

// AuthRequiredMiddleware 强制登录
func AuthRequiredMiddleware(manager service.JwtManager, cmdable redis.Cmdable) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken := utils.GetValueFromCookie(ctx, handler.AccessTokenInCookie)   // 获取 AccessToken
		refreshToken := utils.GetValueFromCookie(ctx, handler.RefreshTokenInCookie) // 获取 RefreshToken

		// 1. 尝试直接通过 AccessToken 认证
		claim, err := manager.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claim != nil {
			// 查询缓存
			ok, err := cmdable.Exists(ctx, service.ClearTokenPrefix+claim.SSid).Result()
			if err != nil || ok == 1 {
				ctx.SetCookie(handler.RefreshTokenInCookie, "", -1, "/", "localhost", false, true)
				ctx.SetCookie(handler.AccessTokenInCookie, "", -1, "/", "localhost", false, true)

				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			// AccessToken 认证直接通过
			slog.Info("AuthMiddleware 认证 AccessToken 成功 ...", "user_id", claim.ID)
			ctx.Set(handler.UserIDInCtx, claim.ID) // 把用户 ID 放入上下文, 以便后续中间件直接使用
			ctx.Next()
			return
		}

		slog.Info("AuthMiddleware 认证 AccessToken 失败, 尝试认证 RefreshToken ...")

		// AccessToken 认证不通过, 尝试通过 RefreshToken 认证
		result := cmdable.Get(ctx, service.FreshTokenPrefix+refreshToken) // 从 Redis 尝试获取 AccessToken
		if result.Err() != nil {
			// 没拿到 redis 中存的 AccessToken, 说明 RefreshToken 也认证不通过, 没招了
			slog.Info("AuthMiddleware 认证 RefreshToken 失败, 需重新登录 ...")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 如果 redis 能拿到, 检验一下是否被踢出, 重新放到 Cookie 中
		newAccessToken := result.Val()
		newClaim, err := manager.VerifyToken(newAccessToken)
		if newClaim != nil {
			// 查询缓存
			ok, err := cmdable.Exists(ctx, service.ClearTokenPrefix+newClaim.SSid).Result()
			if err != nil || ok == 1 {
				ctx.SetCookie(handler.RefreshTokenInCookie, "", -1, "/", "localhost", false, true)
				ctx.SetCookie(handler.AccessTokenInCookie, "", -1, "/", "localhost", false, true)

				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			// 没被踢出, 重放到 Cookie 中
			ctx.SetCookie(handler.AccessTokenInCookie, accessToken, 0, "/", "localhost", false, true)

			slog.Info("AuthMiddleware 认证 RefreshToken 成功 ...", "user_id", newClaim.ID)
			ctx.Set(handler.UserIDInCtx, newClaim.ID) // 把用户 ID 放入上下文, 以便后续中间件直接使用
			ctx.Next()
			return
		}
	}
}
