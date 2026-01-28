package infra

import (
	"context"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/lottery/infra/viper"
)

var (
	globalRedisClient *redis.Client
	redisOnce         sync.Once
)

// Init 连接到 Redis 数据库, 生成一个 *redis.Client 赋给全局数据库变量 globalRedisClient
func Init(confDir, confFileName, confFileType string) redis.UniversalClient {
	// 初始化 Viper 进行配置读取
	viper := viper.InitViper(confDir, confFileName, confFileType)
	host := viper.GetString("redis.host")
	port := viper.GetString("redis.port")
	db := viper.GetInt("redis.db")

	redisAddr := host + ":" + port // 拼接地址
	redisOption := &redis.Options{
		Addr: redisAddr,
		DB:   db,
	}

	// 连接到数据库
	redisOnce.Do(func() {
		globalRedisClient = redis.NewClient(redisOption)
	})

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
