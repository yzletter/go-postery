package middleware

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/handler"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
)

// AuthRequiredMiddleware 强制登录
func AuthRequiredMiddleware(authSvc service.AuthService, cmdable redis.Cmdable) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken := ctx.GetHeader(conf.AccessTokenInHeader)                   // 获取 AccessToken
		refreshToken := utils.GetValueFromCookie(ctx, conf.RefreshTokenInCookie) // 获取 RefreshToken

		// 尝试直接通过 AccessToken 认证
		claim, err := authSvc.VerifyToken(accessToken)
		if err == nil && claim != nil {
			// 黑名单检查
			ssid := claim.SSid
			if ssid == "" {
				unauthorized(ctx)
				return
			}
			ok, err := cmdable.Exists(ctx, conf.ClearTokenPrefix+ssid).Result()
			if err != nil || ok > 0 {
				unauthorized(ctx)
				return
			}

			// AccessToken 认证通过
			slog.Info("AuthMiddleware 认证 AccessToken 成功 ...", "user_id", claim.Uid)
			ctx.Set(handler.UserIDInCtx, claim.Uid) // 把用户 ID 放入上下文, 以便后续中间件直接使用
			ctx.Next()
			return
		}

		if refreshToken == "" { // RefreshToken 不存在
			unauthorized(ctx)
			return
		}

		// 从缓存中获取信息
		mp, err := cmdable.HGetAll(ctx, conf.RefreshTokenPrefix+refreshToken).Result()
		if err != nil || len(mp) == 0 {
			// RefreshToken 已经过期
			unauthorized(ctx)
			return
		}

		id, err1 := strconv.ParseInt(mp["user_id"], 10, 64)
		role, err2 := strconv.Atoi(mp["role"])
		ssid := mp["ssid"]
		if ssid == "" || err1 != nil || err2 != nil {
			unauthorized(ctx)
			return
		}

		// 黑名单检查
		ok, err := cmdable.Exists(ctx, conf.ClearTokenPrefix+ssid).Result()
		if err != nil || ok > 1 {
			unauthorized(ctx)
			return
		}

		// redis 中清除旧 token
		_ = authSvc.ClearTokens(ctx, accessToken, refreshToken)

		// 重新签发 新token
		newAccessToken, newRefreshToken, err := authSvc.IssueTokens(ctx, id, role, ctx.Request.UserAgent())
		if err != nil {
			unauthorized(ctx)
			return
		}

		// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
		setTokens(ctx, newAccessToken, newRefreshToken)

		slog.Info("AuthMiddleware 认证 RefreshToken 成功 ...", "user_id", id)
		ctx.Set(handler.UserIDInCtx, id) // 把用户 ID 放入上下文, 以便后续中间件直接使用
		ctx.Next()
		return
	}
}

func setTokens(ctx *gin.Context, accessToken, refreshToken string) {
	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	ctx.Header(conf.AccessTokenInHeader, accessToken)
	ctx.SetCookie(conf.RefreshTokenInCookie, refreshToken, conf.RefreshTokenCookieMaxAgeSecs, "/", "localhost", false, true)
}

func unauthorized(ctx *gin.Context) {
	// 清除 token
	ctx.Header(conf.AccessTokenInHeader, "")
	ctx.SetCookie(conf.RefreshTokenInCookie, "", -1, "/", "localhost", false, true)
	// 退出
	ctx.AbortWithStatus(http.StatusUnauthorized)
}
