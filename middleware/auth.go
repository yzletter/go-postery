package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/handler"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
)

// AuthRequiredMiddleware 强制登录中间件
func AuthRequiredMiddleware(authSvc service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取 AccessToken 和  RefreshToken
		accessToken := handler.ExtractToken(ctx)
		refreshToken := utils.GetValueFromCookie(ctx, conf.RefreshTokenInCookie)

		slog.Info("获取当前用户的 Token", "AccessToken", accessToken, "RefreshToken", refreshToken)

		// 先尝试通过 AccessToken 认证
		if uid, ok := tryAuthByAccessToken(ctx, authSvc, accessToken); ok {
			accept(ctx, uid)
			return
		}

		// 再尝试通过 RefreshToken 认证
		if uid, ok := tryAuthByRefreshToken(ctx, authSvc, accessToken, refreshToken); ok {
			accept(ctx, uid)
			return
		}

		// 校验失败
		unauthorized(ctx)
		return
	}
}

// 尝试通过 AccessToken 认证, 认证成功返回 uid 和 True
func tryAuthByAccessToken(ctx *gin.Context, authSvc service.AuthService, accessToken string) (int64, bool) {
	if accessToken == "" {
		return 0, false
	}

	claim, err := authSvc.VerifyAccessToken(accessToken)
	if err != nil || claim == nil || claim.SSid == "" {
		return 0, false
	}

	// 黑名单检查
	if ok := isBlacklisted(ctx, authSvc, claim.SSid); ok {
		// 不用再走其他认证了
		unauthorized(ctx)
		return 0, false
	}

	// AccessToken 认证通过
	slog.Info("AuthMiddleware 认证 AccessToken 成功 ...", "user_id", claim.Uid)

	return claim.Uid, true
}

// 尝试通过 RefreshToken 认证, 认证成功返回 uid 和 True
func tryAuthByRefreshToken(ctx *gin.Context, authSvc service.AuthService, accessToken, refreshToken string) (int64, bool) {
	// RefreshToken 不存在
	if refreshToken == "" {
		return 0, false
	}

	// RefreshToken 存在
	uid, role, ssid, err := authSvc.GetInfoByRefreshToken(ctx, refreshToken) // 根据 RefreshToken 从缓存中获取信息
	if err != nil || ssid == "" {
		return 0, false
	}

	// 黑名单检查
	if ok := isBlacklisted(ctx, authSvc, ssid); ok {
		// 不用再走其他认证了
		unauthorized(ctx)
		return 0, false
	}

	// 删除旧 Token
	_ = authSvc.ClearTokens(ctx, accessToken, refreshToken)

	// 重新签发 Token
	newAccessToken, newRefreshToken, err := authSvc.IssueTokens(ctx, uid, role, ctx.Request.UserAgent())
	if err != nil {
		return 0, false
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	setTokens(ctx, newAccessToken, newRefreshToken)

	// RefreshToken 认证通过
	slog.Info("AuthMiddleware 认证 RefreshToken 成功 ...", "user_id", uid)

	return uid, true
}

// 判断是否被拉黑
func isBlacklisted(ctx context.Context, authSvc service.AuthService, ssid string) bool {
	if ssid == "" {
		return true
	}
	exist, err := authSvc.CheckBlackList(ctx, ssid)
	return err != nil || exist // 有错误或者黑名单中存在, 都视为已被拉黑
}

// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
func setTokens(ctx *gin.Context, accessToken, refreshToken string) {
	ctx.Header("Authorization", "Bearer "+accessToken)
	ctx.SetCookie(conf.RefreshTokenInCookie, refreshToken, conf.RefreshTokenMaxAgeSecs, "/", "localhost", false, true)
}

// 返回没有权限的响应
func unauthorized(ctx *gin.Context) {
	// 清除 token
	ctx.Header("Authorization", "")
	ctx.SetCookie(conf.RefreshTokenInCookie, "", -1, "/", "localhost", false, true)

	// 退出
	ctx.AbortWithStatus(http.StatusUnauthorized)
}

// 把用户 ID 放入上下文, 以便后续中间件直接使用
func accept(ctx *gin.Context, uid int64) {
	ctx.Set(handler.UserIDInContext, uid)
	ctx.Next()
}
