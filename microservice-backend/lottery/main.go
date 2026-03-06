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
	"github.com/yzletter/go-postery/microservice-backend/lottery/config"
	grpc_server "github.com/yzletter/go-postery/microservice-backend/lottery/grpc"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/lottery/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/lottery/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/lottery/infra/jaeger"
	infraMySQL "github.com/yzletter/go-postery/microservice-backend/lottery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/lottery/infra/redis"
	infraRocketMQ "github.com/yzletter/go-postery/microservice-backend/lottery/infra/rocketmq"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/lottery/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/lottery/infra/snowflake"
	"github.com/yzletter/go-postery/microservice-backend/lottery/repository"
	"github.com/yzletter/go-postery/microservice-backend/lottery/repository/cache"
	"github.com/yzletter/go-postery/microservice-backend/lottery/repository/dao"
	"github.com/yzletter/go-postery/microservice-backend/lottery/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const ServiceName = "lottery_service"

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{"172.16.131.223:2379"})       // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient, ServiceName+"_") // Get Config From Remote Config Center
	fmt.Printf("%s Init Config Success %+v\n", ServiceName, Config)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(Config.Log)                                            // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, ServiceName) // Init JaegerTracer

	// Infrastructure 层
	RedisClient := infraRedis.Init(Config.Redis)        // Init Redis
	MySQLGormDB := infraMySQL.Init(Config.MySQL)        // Init MySQL
	RocketMQ := infraRocketMQ.Init(Config.RocketMQ)     // Init RocketMQ
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0) // Init Snowflake

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
		AddFunc(infraRedis.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		Build()

	// Start gRPC Server
	lis, err := net.Listen("tcp", Config.GRPC.Addr)
	if err != nil {
		panic(err)
	}
	if err := server.Serve(lis); err != nil {
		slog.Error("Service gRPC Server Start Failed", "service", ServiceName, "error", err)
		panic(err)
	}
}
