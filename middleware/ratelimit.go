package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/service/ratelimit"
)

const RateLimitPrefix = "ip-limit"

func RateLimitMiddleware(rateLimitService *ratelimit.RateLimitService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 传入 Redis 的前缀和当前 IP
		limited, err := rateLimitService.Limit(ctx, RateLimitPrefix, ctx.ClientIP())
		if err != nil {
			slog.Error("RateLimit Wrong", "error", err)
			// 限流出错了 (一般为 Redis 出错)
			// 激进做法: 虽然 Redis 崩溃了, 但为了不影响用户体验，直接放行
			// 保守做法: 由于借助 Redis 进行限流, Redis 崩溃了, 为了防止系统崩溃，直接限流
			ctx.AbortWithStatus(http.StatusInternalServerError) // 这里采用保守做法
			return
		}

		// 需要限流
		if limited {
			slog.Error("对当前 IP 执行限流 ...", "IP", ctx.ClientIP())
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		ctx.Next()
	}
}
