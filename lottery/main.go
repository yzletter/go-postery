package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	auth_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/lottery/config"
	grpc_server "github.com/yzletter/go-postery/lottery/grpc"
	infraEtcd "github.com/yzletter/go-postery/lottery/infra/etcd"
	"github.com/yzletter/go-postery/lottery/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/lottery/infra/jaeger"
	infraMySQL "github.com/yzletter/go-postery/lottery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/lottery/infra/redis"
	infraRocketMQ "github.com/yzletter/go-postery/lottery/infra/rocketmq"
	infraSlog "github.com/yzletter/go-postery/lottery/infra/slog"
	"github.com/yzletter/go-postery/lottery/infra/snowflake"
	"github.com/yzletter/go-postery/lottery/repository"
	"github.com/yzletter/go-postery/lottery/repository/cache"
	"github.com/yzletter/go-postery/lottery/repository/dao"
	"github.com/yzletter/go-postery/lottery/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const ConfigPrefix = "lottery_service_"

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	//EtcdClient := infraEtcd.Init([]string{"172.16.131.223:2379"})    // Init Etcd
	EtcdClient := infraEtcd.Init([]string{"localhost:12379"}) // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient, ConfigPrefix)
	fmt.Printf("Auth Service Init Config Success \n%+v\n", Config)

	// Infra 层
	infraSlog.InitSlog(Config.Log)                                               // 初始化 slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, "auth-service") // Init JaegerTracer
	RedisClient := infraRedis.Init(Config.Redis)                                 // Init Redis
	MySQLGormDB := infraMySQL.Init(Config.MySQL)                                 // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)                          // 初始化 雪花算法
	RocketMQ := infraRocketMQ.Init(Config.RocketMQ)                              // 初始化 RocketMQ

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
	MetricService := service.NewMetricService()
	LotteryService := service.NewLotteryService(OrderRepo, GiftRepo, RocketMQ, IDGenerator) // 注册 LotteryService
	LotteryService.InitCacheInventory(context.Background())
	go LotteryService.StartLotteryOrderConsumer(context.Background()) // 开启协程核查临时订单进行库存回流

	// gRPC Server
	LotteryServiceServer := grpc_server.NewLotteryServiceServer(LotteryService)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	auth_grpc.RegisterLotteryServiceServer(server, LotteryServiceServer) // 注册服务

	// Prometheus
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(Config.Metric.Addr, mux); err != nil {
			slog.Error("Metric Server Failed", "error", err)
		}
	}()

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(infraMySQL.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		Build()

	// 监听本地端口
	lis, err := net.Listen("tcp", Config.GRPC.Addr)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(lis); err != nil {
		slog.Error("Auth grpc Server Start Failed", "error", err)
		panic(err)
	}
}
