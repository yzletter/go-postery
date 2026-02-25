package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	conf2 "github.com/yzletter/go-postery/auth/conf"
	"github.com/yzletter/go-postery/auth/config"
	"github.com/yzletter/go-postery/auth/infra/graceful_stop"
	infraMySQL "github.com/yzletter/go-postery/auth/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/auth/infra/redis"
	"github.com/yzletter/go-postery/auth/infra/security"
	infraSlog "github.com/yzletter/go-postery/auth/infra/slog"
	"github.com/yzletter/go-postery/auth/infra/snowflake"
	"github.com/yzletter/go-postery/auth/repository"
	"github.com/yzletter/go-postery/auth/repository/cache"
	"github.com/yzletter/go-postery/auth/repository/dao"
	"github.com/yzletter/go-postery/auth/service"
	infraEtcd "github.com/yzletter/go-postery/code/infra/etcd"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{"172.16.150.246:2379"}) // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient)

	// Infra 层
	infraSlog.InitSlog(Config.Log)                        // Init Slog
	RedisClient := infraRedis.Init(Config.Redis)          // Init Redis
	MySQLGormDB := infraMySQL.Init(Config.MySQL)          // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)   // 初始化 雪花算法
	PasswordHasher := security.NewBcryptPasswordHasher(0) // 初始化 密码哈希器
	JwtManager := security.NewJwtManager(conf2.JwtTokenKey)
	// Cache 层
	AuthCache := cache.NewAuthCache(RedisClient)
	// DAO 层
	AuthDAO := dao.NewAuthDAO(MySQLGormDB)
	// Repository 层
	AuthRepo := repository.NewAuthRepository(AuthDAO, AuthCache)
	// Service 层
	AuthService := service.NewAuthService(AuthRepo, JwtManager, PasswordHasher, IDGenerator) // 注册 AuthService
	MetricService := service.NewMetricService()

	// gRPC Server
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()),
	)
	auth_grpc.RegisterAuthServiceServer(server, AuthService) // 注册服务

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
		AddFunc(infraRedis.Close).AddFunc(infraMySQL.Close).AddFunc(cancel).
		Build()

	// 监听本地端口
	lis, err := net.Listen("tcp", "localhost:"+Config.GRPC.Port)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(lis); err != nil {
		slog.Error("Auth grpc Server Start Failed", "error", err)
		panic(err)
	}
}
