package service

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/slide_window_script.lua
var luaSlideWindowScript string // luaSlideWindowScript 滑动窗口算法 lua 脚本

type RateLimitService struct {
	redisClient redis.UniversalClient // 依赖 Redis 数据库
	interval    time.Duration         // 窗口大小
	rate        int                   // 阈值
}

func NewRateLimitService(redisClient redis.UniversalClient, interval time.Duration, rate int) *RateLimitService {
	return &RateLimitService{
		redisClient: redisClient,
		interval:    interval,
		rate:        rate,
	}
}

func (svc *RateLimitService) Limit(ctx context.Context, prefix, identifier string) (bool, error) {
	// 拼接 Redis Key
	redisKey := fmt.Sprintf("%s:%s", prefix, identifier)

	// 执行 lua 脚本需要的参数
	windowScale := svc.interval.Milliseconds()
	maxRate := svc.rate
	nowTime := time.Now().UnixMilli()
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	// 返回脚本执行结果
	return svc.redisClient.Eval(ctx, luaSlideWindowScript, []string{redisKey}, windowScale, maxRate, nowTime, requestID).Bool()
}
