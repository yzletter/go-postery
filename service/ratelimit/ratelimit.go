package ratelimit

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed slide_window_script.lua
var luaSlideWindowScript string // luaSlideWindowScript 滑动窗口算法 lua 脚本

type RateLimitService struct {
	redisClient redis.Cmdable // 依赖 Redis 数据库
	internal    time.Duration // 窗口大小
	rate        int           // 阈值
}

func NewRateLimitService(redisClient redis.Cmdable, interval time.Duration, rate int) *RateLimitService {
	return &RateLimitService{
		redisClient: redisClient,
		internal:    interval,
		rate:        rate,
	}
}

func (svc *RateLimitService) Limit(prefix, IP string) (bool, error) {
	// 拼接 Redis Key
	redisKey := fmt.Sprintf("%s:%s", prefix, IP)

	// 执行 lua 脚本需要的参数
	windowScale := svc.internal.Milliseconds()
	maxRate := svc.rate
	nowTime := time.Now().Unix()

	// 返回脚本执行结果
	return svc.redisClient.Eval(luaSlideWindowScript, []string{redisKey}, windowScale, maxRate, nowTime).Bool()
}
