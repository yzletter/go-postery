package infra

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/microservice-backend/lottery/conf"
)

var (
	globalRedisClient *redis.Client
	redisOnce         sync.Once
)

// Init 连接到 Redis 数据库, 生成一个 *redis.Client 赋给全局数据库变量 globalRedisClient
func Init(config conf.RedisConfig) redis.UniversalClient {
	redisOption := &redis.Options{
		Addr: config.Addr,
		DB:   config.DB,
	}

	// 连接到数据库
	redisOnce.Do(func() {
		globalRedisClient = redis.NewClient(redisOption)
	})

	fmt.Println(config.Addr, config.DB)

	// 尝试 ping 通
	if err := globalRedisClient.Ping(context.Background()).Err(); err != nil { // 须加上.Err(), 否则会报 ping 通错
		slog.Error("初始化 Redis 失败 ...", "error", err)
		panic(err)
	} else {
		slog.Info("初始化 Redis 成功 ...")
	}

	return globalRedisClient
}

// Ping ping 一下数据库 保持连接
func Ping() {
	if globalRedisClient != nil {
		err := globalRedisClient.Ping(context.Background()).Err()
		if err != nil {
			slog.Info("Ping Redis 失败 ...")
			return
		}
		slog.Info("Ping Redis 成功 ...")
		return
	}
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
