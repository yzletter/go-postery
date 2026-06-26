package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	"github.com/yzletter/go-postery/backend/bff/handler"
	"github.com/yzletter/go-postery/backend/conf"
	grpcclient "github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/utils"
)

// AuthRequiredMiddleware 强制登录中间件
func AuthRequiredMiddleware(client grpcclient.AuthClient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取 AccessToken 和  RefreshToken
		accessToken := handler.ExtractToken(ctx)
		refreshToken := utils.GetValueFromCookie(ctx, conf.RefreshTokenInCookie)

		slog.Info("获取当前用户的 Token", "AccessToken", maskTokenForLog(accessToken), "RefreshToken", maskTokenForLog(refreshToken))

		// 先尝试通过 AccessToken 认证
		if uid, ok := tryAuthByAccessToken(ctx, client, accessToken); ok {
			accept(ctx, uid)
			return
		}

		// 再尝试通过 RefreshToken 认证
		if uid, ok := tryAuthByRefreshToken(ctx, client, accessToken, refreshToken); ok {
			accept(ctx, uid)
			return
		}

		// 校验失败
		unauthorized(ctx)
		return
	}
}

// 尝试通过 AccessToken 认证, 认证成功返回 uid 和 True
func tryAuthByAccessToken(ctx *gin.Context, client grpcclient.AuthClient, accessToken string) (int64, bool) {
	if accessToken == "" {
		return 0, false
	}

	claim, err := client.VerifyAccessToken(ctx, &auth_grpc.AccessToken{AccessToken: accessToken})
	if err != nil || claim == nil || claim.SSID == "" {
		return 0, false
	}

	// 黑名单检查
	if ok := isBlacklisted(ctx, client, claim.SSID); ok {
		// 不用再走其他认证了
		unauthorized(ctx)
		return 0, false
	}

	// AccessToken 认证通过
	slog.Info("AuthMiddleware 认证 AccessToken 成功 ...", "user_id", claim.UserID)

	return claim.UserID, true
}

// 尝试通过 RefreshToken 认证, 认证成功返回 uid 和 True
func tryAuthByRefreshToken(ctx *gin.Context, client grpcclient.AuthClient, accessToken, refreshToken string) (int64, bool) {
	// RefreshToken 不存在
	if refreshToken == "" {
		return 0, false
	}

	// RefreshToken 存在
	userInfo, err := client.GetInfoByRefreshToken(ctx, &auth_grpc.RefreshToken{RefreshToken: refreshToken}) // 根据 RefreshToken 从缓存中获取信息
	if err != nil || userInfo.SSID == "" {
		return 0, false
	}

	// 黑名单检查
	if ok := isBlacklisted(ctx, client, userInfo.SSID); ok {
		// 不用再走其他认证了
		unauthorized(ctx)
		return 0, false
	}

	// 删除旧 Token
	_, _ = client.ClearTokens(ctx, &auth_grpc.DualTokens{AccessToken: accessToken, RefreshToken: refreshToken})

	// 重新签发 Token
	tokens, err := client.IssueTokens(ctx, &auth_grpc.IssueTokenRequest{UserID: userInfo.UserID, Role: userInfo.Role, UserAgent: ctx.Request.UserAgent()})
	if err != nil {
		return 0, false
	}

	// 将 AccessToken 放进 Header, RefreshToken 放进 Cookie
	setTokens(ctx, tokens.AccessToken, tokens.RefreshToken)

	// RefreshToken 认证通过
	slog.Info("AuthMiddleware 认证 RefreshToken 成功 ...", "user_id", userInfo.UserID)

	return userInfo.UserID, true
}

// 判断是否被拉黑
func isBlacklisted(ctx context.Context, client grpcclient.AuthClient, ssid string) bool {
	if ssid == "" {
		return true
	}
	resp, err := client.CheckBlackList(ctx, &auth_grpc.CheckBlackListRequest{SSID: ssid})
	return err != nil || resp.Result // 有错误或者黑名单中存在, 都视为已被拉黑
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
	ctx.Set(conf.UserIDInContext, uid)
	ctx.Next()
}

func maskTokenForLog(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 12 {
		return "***"
	}
	return token[:6] + "..." + token[len(token)-6:]
}
