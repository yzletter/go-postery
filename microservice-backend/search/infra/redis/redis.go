package infra

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/microservice-backend/search/config"
)

var (
	globalRedisClient *redis.Client
	redisOnce         sync.Once
)

// Init 连接到 Redis 数据库, 生成一个 *redis.Client 赋给全局数据库变量 globalRedisClient
func Init(config config.RedisConfig) redis.UniversalClient {
	redisOption := &redis.Options{
		Addr: config.Addr,
		DB:   config.DB,
	}

	redisOnce.Do(func() {
		globalRedisClient = redis.NewClient(redisOption)
	})

	fmt.Println(config.Addr, config.DB)

	if err := globalRedisClient.Ping(context.Background()).Err(); err != nil {
		slog.Error("初始化 Redis 失败 ...", "error", err)
		panic(err)
	} else {
		slog.Info("初始化 Redis 成功 ...")
	}

	return globalRedisClient
}

func Close() {
	if globalRedisClient != nil {
		err := globalRedisClient.Close()
		if err != nil {
			slog.Info("关闭 Redis 失败 ...")
			return
		}
		slog.Info("关闭 Redis 成功 ...")
		return
	}
}
