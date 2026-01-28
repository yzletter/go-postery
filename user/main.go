package main

import (
	"log/slog"
	"net"

	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	infraKafka "github.com/yzletter/go-postery/outbox/infra/kafka"
	"github.com/yzletter/go-postery/user/conf"
	infraMySQL "github.com/yzletter/go-postery/user/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/user/infra/redis"
	infraSlog "github.com/yzletter/go-postery/user/infra/slog"
	"github.com/yzletter/go-postery/user/infra/snowflake"
	"github.com/yzletter/go-postery/user/infra/viper"
	"github.com/yzletter/go-postery/user/repository"
	"github.com/yzletter/go-postery/user/repository/cache"
	"github.com/yzletter/go-postery/user/repository/dao"
	"github.com/yzletter/go-postery/user/service"
	"google.golang.org/grpc"
)

func main() {
	// Infra 层
	infraSlog.InitSlog(conf.LogFilePath)                                                                           // 初始化 slog
	RedisClient := infraRedis.Init("./conf", "cache", viper.YAML)                                                  // 初始化 Redis
	MySQLGormDB := infraMySQL.Init("./conf", "db", viper.YAML, "./logs")                                           // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)                                                            // 初始化 雪花算法
	FollowKafkaConsumer := infraKafka.InitConsumer([]string{conf.KafkaEndpoint}, conf.KafkaTopic, conf.KafkaGroup) // 初始化 Follow 模块 Kafka 消费方
	// Cache 层
	UserCache := cache.NewUserCache(RedisClient)
	// DAO 层
	UserDAO := dao.NewUserDAO(MySQLGormDB)
	FollowDAO := dao.NewFollowDAO(MySQLGormDB)
	// Repository 层
	UserRepo := repository.NewUserRepository(UserDAO, UserCache) // 注册 userRepo
	FollowRepo := repository.NewFollowRepository(FollowDAO)      // 注册 FollowRepository
	// Service 层
	UserService := service.NewUserService(UserRepo, FollowRepo, FollowKafkaConsumer, IDGenerator) // 注册 userSvc

	// 监听本地端口
	lis, err := net.Listen("tcp", "localhost:"+conf.Port)
	if err != nil {
		panic(err)
	}

	// 启动 grpc 服务
	server := grpc.NewServer()
	user_grpc.RegisterUserServiceServer(server, UserService) // 注册服务
	if err := server.Serve(lis); err != nil {
		slog.Error("User grpc Server Start Failed", "error", err)
		panic(err)
	}
}
