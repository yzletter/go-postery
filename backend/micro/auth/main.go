package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	"github.com/yzletter/go-postery/backend/conf"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/database/mysql"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	"github.com/yzletter/go-postery/backend/infra/security"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/snowflake"
	"github.com/yzletter/go-postery/backend/micro/auth/config"
	server "github.com/yzletter/go-postery/backend/micro/auth/grpc"
	"github.com/yzletter/go-postery/backend/micro/auth/repository"
	"github.com/yzletter/go-postery/backend/micro/auth/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/auth/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/auth/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Service           = manager.AuthService // 微服务名
	GoPostery         = conf.GoPostery      // GoPostery
	CommonConfPrefix  = GoPostery + "/conf/common_conf/"
	ServiceConfPrefix = GoPostery + "/conf/service_conf/" + Service + "_conf/"
)

func main() {
	// 启动参数, etcd 地址
	etcdEndpoint := flag.String("etcd", "localhost:2379", "etcd 地址")
	flag.Parse()

	// 获取本地内网 IP
	ip, err := utils.GetLocalIP()
	if err != nil {
		slog.Error("get local IP failed", "error", err)
		panic(err)
	}

	// 全局 Context
	ctx, cancel := context.WithCancel(context.Background())

	// Init Etcd
	etcdClient := infraEtcd.Init([]string{*etcdEndpoint})

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, etcdClient, CommonConfPrefix)
	// 加载私有配置
	AuthServiceConf := config.LoadAuthServiceConfig(ctx, etcdClient, ServiceConfPrefix)

	// Infrastructure 层
	infraSlog.InitSlog(AuthServiceConf.Log) // Init Slog

	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service) // Init JaegerTracer
	RedisClient := infraRedis.Init(CommonMicroConf.Redis)                          // Init Redis
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL)                          // Init MySQL
	JwtManager := security.NewJwtManager(conf.JwtTokenKey)                         // Init JWT
	PasswordHasher := security.NewBcryptPasswordHasher(0)                          // Init PasswordHasher
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)                            // Init Snowflake

	// Cache 层
	AuthCache := cache.NewAuthCache(RedisClient)
	// DAO 层
	AuthDAO := dao.NewAuthDAO(MySQLGormDB)
	// Repository 层
	AuthRepo := repository.NewAuthRepository(AuthDAO, AuthCache)

	// ServiceHub 服务发现中心
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, etcdClient, hub.NewRoundRobinLoadBalancer())

	// gRPC Client
	CodeManager := manager.NewCodeManager(ctx, manager.CodeService, ETCDServiceHub)

	// Service 层
	AuthService := service.NewAuthService(AuthRepo, JwtManager, PasswordHasher, IDGenerator, CodeManager)
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 1000)
	MetricService := pkg.NewMetricService(Service)

	// gRPC Server
	AuthServiceServer := server.NewAuthServiceServer(AuthService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	auth_grpc.RegisterAuthServiceServer(ServiceRegistrar, AuthServiceServer) // 注册微服务

	// Prometheus
	metricAddr := ip + ":" + AuthServiceConf.Metric.Port
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("Metric Server Failed", "error", err)
		}
	}()

	grpcAddr := ip + ":" + AuthServiceConf.GRPC.Port
	if lis, err := net.Listen("tcp", grpcAddr); err != nil {
		panic(err)
	} else {
		go func() {
			if err := ServiceRegistrar.Serve(lis); err != nil {
				slog.Error("Service gRPC Server Start Failed", "service", Service, "error", err)
				panic(err)
			}
		}()
	}

	slog.Info("ready to register", "service", Service, "grpc_addr", grpcAddr, "metric_addr", metricAddr)

	// 向服务发现中心注册服务
	leaseID, err := ETCDServiceHub.Register(ctx, Service, grpcAddr, 0)
	if err != nil {
		slog.Error("Service Auth Server Register Failed", "service", Service, "error", err)
		panic(err)
	}

	// 向服务发现中心自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, Service, grpcAddr, leaseID)
			if err != nil {
				slog.Error("Service Auth Server Register Failed", "service", Service, "error", err)
			}
			time.Sleep(time.Duration(CommonMicroConf.ServiceHub.HeartbeatFrequency)*time.Second - 200*time.Millisecond)
		}
	}()

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(infraMySQL.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		AddFunc(func() {
			// 注销服务
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := ETCDServiceHub.Unregister(ctx, Service, grpcAddr); err != nil {
				slog.Error("Service Auth Server Unregister Failed", "service", Service, "error", err)
			}
		}).
		BuildBlock()
}
