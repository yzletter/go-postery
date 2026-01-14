package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/handler"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
)

// AuthRequiredMiddleware 强制登录
func AuthRequiredMiddleware(authSvc service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken := handler.ExtractToken(ctx)                                 // 获取 AccessToken
		refreshToken := utils.GetValueFromCookie(ctx, conf.RefreshTokenInCookie) // 获取 RefreshToken

		slog.Info("获取当前用户的 Token", "AccessToken", accessToken, "RefreshToken", refreshToken)

		// 尝试直接通过 AccessToken 认证
		claim, err := authSvc.VerifyAccessToken(accessToken)
		if err == nil && claim != nil {
			// 黑名单检查
			ssid := claim.SSid
			if ssid == "" {
				unauthorized(ctx)
				return
			}

			exist, err := authSvc.CheckBlackList(ctx, ssid) // 判断在黑名单中是否存在
			if err != nil || exist {
				unauthorized(ctx)
				return
			}

			// AccessToken 认证通过
			slog.Info("AuthMiddleware 认证 AccessToken 成功 ...", "user_id", claim.Uid)

			// 把用户 ID 放入上下文, 以便后续中间件直接使用
			ctx.Set(handler.UserIDInContext, claim.Uid)
			ctx.Next()
			return
		}

		// RefreshToken 不存在
		if refreshToken == "" {
			unauthorized(ctx)
			return
		}

		// RefreshToken 存在, 根据 RefreshToken 从缓存中获取信息
		uid, role, ssid, err := authSvc.GetInfoByRefreshToken(ctx, refreshToken)
		if err != nil {
			unauthorized(ctx)
			return
		}

		// 黑名单检查
		exist, err := authSvc.CheckBlackList(ctx, ssid) // 判断在黑名单中是否存在
		if err != nil || exist {
			unauthorized(ctx)
			return
		}

		// 删除旧 Tokens
		_ = authSvc.ClearTokens(ctx, accessToken, refreshToken)

		// 重新签发 新token
		newAccessToken, newRefreshToken, err := authSvc.IssueTokens(ctx, uid, role, ctx.Request.UserAgent())
		if err != nil {
			unauthorized(ctx)
			return
		}

		// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
		setTokens(ctx, newAccessToken, newRefreshToken)

		slog.Info("AuthMiddleware 认证 RefreshToken 成功 ...", "user_id", uid)

		// 把用户 ID 放入上下文, 以便后续中间件直接使用
		ctx.Set(handler.UserIDInContext, uid)
		ctx.Next()
		return
	}
}

func setTokens(ctx *gin.Context, accessToken, refreshToken string) {
	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	ctx.Header("Authorization", "Bearer "+accessToken)
	ctx.SetCookie(conf.RefreshTokenInCookie, refreshToken, conf.RefreshTokenMaxAgeSecs, "/", "localhost", false, true)
}

func unauthorized(ctx *gin.Context) {
	// 清除 token
	ctx.Header("Authorization", "")
	ctx.SetCookie(conf.RefreshTokenInCookie, "", -1, "/", "localhost", false, true)

	// 退出
	ctx.AbortWithStatus(http.StatusUnauthorized)
}
