package service

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/slide_window_script.lua
var luaSlideWindowScript string

type RateLimitService struct {
	redisClient redis.UniversalClient
	internal    time.Duration
	rate        int
}

func NewRateLimitService(redisClient redis.UniversalClient, interval time.Duration, rate int) *RateLimitService {
	return &RateLimitService{
		redisClient: redisClient,
		internal:    interval,
		rate:        rate,
	}
}

func (svc *RateLimitService) Limit(ctx context.Context, prefix, IP string) (bool, error) {
	redisKey := fmt.Sprintf("%s:%s", prefix, IP)
	windowScale := svc.internal.Milliseconds()
	maxRate := svc.rate
	nowTime := time.Now().UnixMilli()
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	return svc.redisClient.Eval(ctx, luaSlideWindowScript, []string{redisKey}, windowScale, maxRate, nowTime, requestID).Bool()
}
