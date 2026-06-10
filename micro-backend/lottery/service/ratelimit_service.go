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
	redisKey := fmt.Sprintf("%s:%s", prefix, identifier)
	windowScale := svc.interval.Milliseconds()
	maxRate := svc.rate
	nowTime := time.Now().UnixMilli()
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	return svc.redisClient.Eval(ctx, luaSlideWindowScript, []string{redisKey}, windowScale, maxRate, nowTime, requestID).Bool()
}
