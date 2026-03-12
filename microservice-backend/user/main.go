package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/microservice-backend/user/config"
	grpc_server "github.com/yzletter/go-postery/microservice-backend/user/grpc"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/user/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/user/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/user/infra/jaeger"
	infraKafka "github.com/yzletter/go-postery/microservice-backend/user/infra/kafka"
	infraMySQL "github.com/yzletter/go-postery/microservice-backend/user/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/user/infra/redis"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/user/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/user/infra/snowflake"
	repository2 "github.com/yzletter/go-postery/microservice-backend/user/repository"
	"github.com/yzletter/go-postery/microservice-backend/user/repository/cache"
	dao2 "github.com/yzletter/go-postery/microservice-backend/user/repository/dao"
	service2 "github.com/yzletter/go-postery/microservice-backend/user/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const ServiceName = "user_service"

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
	MySQLGormDB := infraMySQL.Init(Config.MySQL)        // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0) // 初始化 雪花算法

	FollowKafkaConsumer := infraKafka.InitConsumer(Config.Kafka) // 初始化 Follow 模块 Kafka 消费方

	// Cache 层
	UserCache := cache.NewUserCache(RedisClient)
	// DAO 层
	UserDAO := dao2.NewUserDAO(MySQLGormDB)
	FollowDAO := dao2.NewFollowDAO(MySQLGormDB)
	// Repository 层
	UserRepo := repository2.NewUserRepository(UserDAO, UserCache) // 注册 userRepo
	FollowRepo := repository2.NewFollowRepository(FollowDAO)      // 注册 FollowRepository
	// Service 层
	UserService := service2.NewUserService(UserRepo, FollowRepo, FollowKafkaConsumer, IDGenerator) // 注册 userSvc
	MetricService := service2.NewMetricService(ServiceName)

	go UserService.StartInitUserScoreConsumer(ctx)

	// gRPC Server
	UserServiceServer := grpc_server.NewUserServiceServer(UserService)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	user_grpc.RegisterUserServiceServer(server, UserServiceServer) // Register gRPC Service

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
