package redis

import (
	"log/slog"
	"sync"

	"github.com/go-redis/redis"
	"github.com/yzletter/go-postery/utils"
)

var (
	GoPosteryRedisClient *redis.Client
	redisOnce            sync.Once
)

// ConnectToRedis 连接到 MySQL 数据库, 生成一个 *redis.Client 赋给全局数据库变量 GoPosteryRedisClient
func ConnectToRedis(confDir, confFileName, confFileType string) {
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
		GoPosteryRedisClient = redis.NewClient(redisOption)
	})

	// 尝试 ping 通
	if err := GoPosteryRedisClient.Ping().Err(); err != nil { // 须加上.Err(), 否则会报 ping 通错
		slog.Error("connect to Redis failed", "error", err)
		panic(err)
	} else {
		slog.Info("connect to Redis succeed")
	}
}

// Ping ping 一下数据库 保持连接
func Ping() {
	if GoPosteryRedisClient != nil {
		err := GoPosteryRedisClient.Ping().Err()
		if err != nil {
			slog.Info("ping GoPosteryRedisClient failed")
			return
		}
		slog.Info("ping GoPosteryRedisClient succeed")
		return
	}
}

func CloseConnection() {
	if GoPosteryRedisClient != nil {
		err := GoPosteryRedisClient.Close()
		if err != nil {
			slog.Info("close GoPosteryRedisClient failed")
			return
		}
		slog.Info("close GoPosteryRedisClient succeed")
		return
	}
}
