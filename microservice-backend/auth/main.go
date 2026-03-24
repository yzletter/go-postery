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
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	"github.com/yzletter/go-postery/microservice-backend/auth/conf"
	"github.com/yzletter/go-postery/microservice-backend/auth/config"
	grpc_server "github.com/yzletter/go-postery/microservice-backend/auth/grpc"
	"github.com/yzletter/go-postery/microservice-backend/auth/grpc/client"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/auth/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/auth/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/auth/infra/jaeger"
	infraMySQL "github.com/yzletter/go-postery/microservice-backend/auth/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/auth/infra/redis"
	"github.com/yzletter/go-postery/microservice-backend/auth/infra/security"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/auth/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/auth/infra/snowflake"
	"github.com/yzletter/go-postery/microservice-backend/auth/repository"
	"github.com/yzletter/go-postery/microservice-backend/auth/repository/cache"
	"github.com/yzletter/go-postery/microservice-backend/auth/repository/dao"
	"github.com/yzletter/go-postery/microservice-backend/auth/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const ServiceName = "auth_service"

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
	RedisClient := infraRedis.Init(Config.Redis)           // Init Redis
	MySQLGormDB := infraMySQL.Init(Config.MySQL)           // Init MySQL
	JwtManager := security.NewJwtManager(conf.JwtTokenKey) // Init JWT
	PasswordHasher := security.NewBcryptPasswordHasher(0)  // Init PasswordHasher
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)    // Init Snowflake

	// Cache 层
	AuthCache := cache.NewAuthCache(RedisClient)
	// DAO 层
	AuthDAO := dao.NewAuthDAO(MySQLGormDB)
	// Repository 层
	AuthRepo := repository.NewAuthRepository(AuthDAO, AuthCache)

	// gRPC Client
	CodeClient, err := client.NewCodeClient()
	if err != nil {
		slog.Error("Init Code gRPC Client Failed", "error", err)
	}

	// Service 层
	AuthService := service.NewAuthService(AuthRepo, JwtManager, PasswordHasher, IDGenerator, CodeClient) // 注册 AuthService
	RateLimitService := service.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := service.NewMetricService(ServiceName)

	// gRPC Server
	AuthServiceServer := grpc_server.NewAuthServiceServer(AuthService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	auth_grpc.RegisterAuthServiceServer(server, AuthServiceServer) // 注册服务

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
		AddFunc(infraRedis.Close).AddFunc(infraMySQL.Close).AddFunc(cancel).AddFunc(TracerShutdown).AddFunc(CodeClient.Close).
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
