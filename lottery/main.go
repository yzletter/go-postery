package main

import (
	"context"
	"log/slog"
	"net"

	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/lottery/conf"
	infraMySQL "github.com/yzletter/go-postery/lottery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/lottery/infra/redis"
	infraRocketMQ "github.com/yzletter/go-postery/lottery/infra/rocketmq"
	infraSlog "github.com/yzletter/go-postery/lottery/infra/slog"
	"github.com/yzletter/go-postery/lottery/infra/snowflake"
	"github.com/yzletter/go-postery/lottery/infra/viper"
	"github.com/yzletter/go-postery/lottery/repository"
	"github.com/yzletter/go-postery/lottery/repository/cache"
	"github.com/yzletter/go-postery/lottery/repository/dao"
	"github.com/yzletter/go-postery/lottery/service"
	"google.golang.org/grpc"
)

func main() {
	// Infra 层
	infraSlog.InitSlog(conf.LogFilePath)                                 // 初始化 slog
	RedisClient := infraRedis.Init("./conf", "cache", viper.YAML)        // 初始化 Redis
	MySQLGormDB := infraMySQL.Init("./conf", "db", viper.YAML, "./logs") // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)                  // 初始化 雪花算法
	RocketMQ := infraRocketMQ.Init(conf.RocketProxyEndpoint)             // 初始化 RocketMQ
	// Cache 层
	GiftCache := cache.NewGiftCache(RedisClient)
	OrderCache := cache.NewOrderCache(RedisClient)
	// DAO 层
	GiftDAO := dao.NewGiftDAO(MySQLGormDB)
	OrderDAO := dao.NewOrderDAO(MySQLGormDB)
	// Repository 层
	GiftRepo := repository.NewGiftRepository(GiftDAO, GiftCache)
	OrderRepo := repository.NewOrderRepository(OrderDAO, OrderCache)
	// Service 层
	LotteryService := service.NewLotteryService(OrderRepo, GiftRepo, RocketMQ, IDGenerator) // 注册 LotteryService
	LotteryService.InitCacheInventory(context.Background())
	go LotteryService.StartLotteryOrderConsumer(context.Background()) // 开启协程核查临时订单进行库存回流

	// 监听本地端口
	lis, err := net.Listen("tcp", "localhost:"+conf.Port)
	if err != nil {
		panic(err)
	}

	// 启动 grpc 服务
	server := grpc.NewServer()
	lottery_grpc.RegisterLotteryServiceServer(server, LotteryService) // 注册服务
	if err := server.Serve(lis); err != nil {
		slog.Error("User grpc Server Start Failed", "error", err)
		panic(err)
	}
}
