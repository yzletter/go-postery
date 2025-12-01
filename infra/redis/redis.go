package infra

import (
	"log/slog"
	"sync"

	"github.com/go-redis/redis"
	"github.com/yzletter/go-postery/utils"
)

var (
	globalRedisClient *redis.Client
	redisOnce         sync.Once
)

// Init 连接到 Redis 数据库, 生成一个 *redis.Client 赋给全局数据库变量 globalRedisClient
func Init(confDir, confFileName, confFileType string) {
	// 初始化 Viper 进行配置读取
	viper := utils.InitViper(confDir, confFileName, confFileType)
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
	if err := globalRedisClient.Ping().Err(); err != nil { // 须加上.Err(), 否则会报 ping 通错
		slog.Error("connect to Redis failed", "error", err)
		panic(err)
	} else {
		slog.Info("connect to Redis succeed")
	}
}

// GetRedis 获取全局DB
func GetRedis() *redis.Client {
	if globalRedisClient == nil {
		slog.Error("未初始化 Redis ...")
	}
	return globalRedisClient
}

// Ping ping 一下数据库 保持连接
func Ping() {
	if globalRedisClient != nil {
		err := globalRedisClient.Ping().Err()
		if err != nil {
			slog.Info("ping globalRedisClient failed")
			return
		}
		slog.Info("ping globalRedisClient succeed")
		return
	}
}

func Close() {
	if globalRedisClient != nil {
		err := globalRedisClient.Close()
		if err != nil {
			slog.Info("close globalRedisClient failed")
			return
		}
		slog.Info("close globalRedisClient succeed")
		return
	}
}
