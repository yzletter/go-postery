package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"github.com/yzletter/go-postery/microservice-backend/session/config"
	grpc_server "github.com/yzletter/go-postery/microservice-backend/session/grpc"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/session/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/session/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/session/infra/jaeger"
	infraKafka "github.com/yzletter/go-postery/microservice-backend/session/infra/kafka"
	infraMySQL "github.com/yzletter/go-postery/microservice-backend/session/infra/mysql"
	infraRabbitMQ "github.com/yzletter/go-postery/microservice-backend/session/infra/rabbitmq"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/session/infra/redis"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/session/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/session/infra/snowflake"
	repository2 "github.com/yzletter/go-postery/microservice-backend/session/repository"
	dao2 "github.com/yzletter/go-postery/microservice-backend/session/repository/dao"
	service2 "github.com/yzletter/go-postery/microservice-backend/session/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const ServiceName = "session_service"

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
	RedisClient := infraRedis.Init(Config.Redis)        // 初始化 Redis
	RabbitMQ := infraRabbitMQ.Init(Config.RabbitMQ)     // 初始化 RabbitMQ
	MySQLGormDB := infraMySQL.Init(Config.MySQL)        // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0) // 初始化 雪花算法

	SessionKafkaConsumer := infraKafka.InitConsumer(Config.Kafka) // 初始化 Session 模块 Kafka 消费方

	// DAO 层
	MessageDAO := dao2.NewMessageDAO(MySQLGormDB)
	SessionDAO := dao2.NewSessionDAO(MySQLGormDB)
	// Repository 层
	MessageRepo := repository2.NewMessageRepository(MessageDAO) // 注册 MessageRepository
	SessionRepo := repository2.NewSessionRepository(SessionDAO) // 注册 SessionRepository
	// Service 层
	SessionService := service2.NewSessionService(SessionRepo, MessageRepo, RabbitMQ, SessionKafkaConsumer, IDGenerator) // 注册 SessionService
	RateLimitService := service2.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := service2.NewMetricService(ServiceName)

	go SessionService.StartSessionRegisterConsumer(ctx) // 开启协程注册新用户聊天功能

	// gRPC Server
	SessionServiceServer := grpc_server.NewSessionServiceServer(SessionService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	session_grpc.RegisterSessionServiceServer(server, SessionServiceServer) // Register gRPC Service

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
