package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	"github.com/yzletter/go-postery/auth/conf"
	"github.com/yzletter/go-postery/auth/config"
	grpc_server "github.com/yzletter/go-postery/auth/grpc"
	"github.com/yzletter/go-postery/auth/grpc/client"
	infraEtcd "github.com/yzletter/go-postery/auth/infra/etcd"
	"github.com/yzletter/go-postery/auth/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/auth/infra/jaeger"
	infraMySQL "github.com/yzletter/go-postery/auth/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/auth/infra/redis"
	"github.com/yzletter/go-postery/auth/infra/security"
	infraSlog "github.com/yzletter/go-postery/auth/infra/slog"
	"github.com/yzletter/go-postery/auth/infra/snowflake"
	"github.com/yzletter/go-postery/auth/repository"
	"github.com/yzletter/go-postery/auth/repository/cache"
	"github.com/yzletter/go-postery/auth/repository/dao"
	"github.com/yzletter/go-postery/auth/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const ConfigPrefix = "auth_service_"

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	//EtcdClient := infraEtcd.Init([]string{"172.16.131.223:2379"})    // Init Etcd
	EtcdClient := infraEtcd.Init([]string{"localhost:12379"}) // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient, ConfigPrefix)
	fmt.Printf("Auth Service Init Config Success \n%+v\n", Config)

	// Infra 层
	infraSlog.InitSlog(Config.Log)                                               // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, "auth-service") // Init JaegerTracer
	RedisClient := infraRedis.Init(Config.Redis)                                 // Init Redis
	MySQLGormDB := infraMySQL.Init(Config.MySQL)                                 // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)                          // 初始化 雪花算法
	PasswordHasher := security.NewBcryptPasswordHasher(0)                        // 初始化 密码哈希器
	JwtManager := security.NewJwtManager(conf.JwtTokenKey)

	// Cache 层
	AuthCache := cache.NewAuthCache(RedisClient)

	// DAO 层
	AuthDAO := dao.NewAuthDAO(MySQLGormDB)

	// Repository 层
	AuthRepo := repository.NewAuthRepository(AuthDAO, AuthCache)
	MetricService := service.NewMetricService()

	// gRPC Client
	CodeClient, err := client.NewCodeClient("localhost:9001")
	if err != nil {
		slog.Error("Init Code gRPC Client Failed", "error", err)
	}

	// Service 层
	AuthService := service.NewAuthService(AuthRepo, JwtManager, PasswordHasher, IDGenerator, CodeClient) // 注册 AuthService

	// gRPC Server
	AuthServiceServer := grpc_server.NewAuthServiceServer(AuthService)
	server := grpc.NewServer(
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
