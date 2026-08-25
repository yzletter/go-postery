package infra

import (
	"context"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/backend/conf"
)

var (
	client *redis.Client
	once   sync.Once
)

// Init 初始化 Redis 数据库
func Init(config conf.RedisConfig) *redis.Client {
	// 连接数据库
	once.Do(func() {
		// Redis 配置
		options := &redis.Options{
			Addr: config.Addr, // 地址
			DB:   0,           // 数据库号
		}

		// 赋值给全局变量
		client = redis.NewClient(options)
	})

	// 尝试 Ping 通, 须加上.Err(), 否则会报 Ping 通错误
	if err := client.Ping(context.Background()).Err(); err != nil {
		slog.Error("init Redis failed ...", "error", err)
		panic(err)
	}

	// 初始化成功
	slog.Info("init Redis success ...")
	return client
}

func Close() {
	if client != nil {
		// 尝试关闭连接
		if err := client.Close(); err != nil {
			slog.Info("close Redis failed ...")
			return
		}

		slog.Info("close Redis success ...")
		return
	}
}
