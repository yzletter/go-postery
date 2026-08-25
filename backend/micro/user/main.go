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
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/conf"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/database/mysql"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/snowflake"
	"github.com/yzletter/go-postery/backend/micro/user/config"
	server "github.com/yzletter/go-postery/backend/micro/user/grpc"
	"github.com/yzletter/go-postery/backend/micro/user/repository"
	"github.com/yzletter/go-postery/backend/micro/user/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/user/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/user/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Service           = manager.UserService // 微服务名
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
	UserServiceConf := config.LoadUserServiceConfig(ctx, etcdClient, ServiceConfPrefix)

	// 公共基础设施
	infraSlog.InitSlog(UserServiceConf.Log)                                        // 初始化日志
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service) // 初始化 Jaeger

	// 基础设施层
	RedisClient := infraRedis.Init(CommonMicroConf.Redis) // 初始化 Redis
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL) // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)   // 初始化雪花算法

	// Cache 层
	UserCache := cache.NewUserCache(RedisClient)
	// DAO 层
	UserDAO := dao.NewUserDAO(MySQLGormDB)
	// Repository 层
	UserRepo := repository.NewUserRepository(UserDAO, UserCache) // 注册 UserRepository

	// 服务注册中心
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, etcdClient, hub.NewRoundRobinLoadBalancer())

	// gRPC Client 层
	InteractiveManager := manager.NewInteractiveManager(ctx, manager.InteractiveService, ETCDServiceHub) // 注册 InteractiveClient
	RankManager := manager.NewRankManager(ctx, manager.RankService, ETCDServiceHub)                      // 注册 RankClient
	OSSManager := manager.NewOSSManager(ctx, manager.OSSService, ETCDServiceHub)                         // 注册 OSSManager

	// Service 层
	UserService := service.NewUserService(UserRepo, InteractiveManager, RankManager, OSSManager, IDGenerator) // 注册 UserService
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 1000)
	MetricService := pkg.NewMetricService(Service) // 注册指标服务

	// gRPC Server 层
	UserServiceServer := server.NewUserServiceServer(UserService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	user_grpc.RegisterUserServiceServer(ServiceRegistrar, UserServiceServer) // 注册 gRPC Service

	// Prometheus 指标服务
	metricAddr := ip + ":" + UserServiceConf.Metric.Port
	go func() {
		mux := http.NewServeMux()
		// 注册指标接口
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("metric server failed", "addr", metricAddr, "error", err)
		}
	}()

	grpcAddr := ip + ":" + UserServiceConf.GRPC.Port
	if lis, err := net.Listen("tcp", grpcAddr); err != nil {
		panic(err)
	} else {
		go func() {
			if err := ServiceRegistrar.Serve(lis); err != nil {
				slog.Error("grpc server failed", "service", Service, "addr", grpcAddr, "error", err)
				panic(err)
			}
		}()
	}

	slog.Info("ready to register", "service", Service, "grpc_addr", grpcAddr, "metric_addr", metricAddr)

	// 向服务发现中心注册服务
	leaseID, err := ETCDServiceHub.Register(ctx, Service, grpcAddr, 0)
	if err != nil {
		slog.Error("register user service failed", "service", Service, "addr", grpcAddr, "error", err)
		panic(err)
	}

	// 向服务发现中心自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, Service, grpcAddr, leaseID)
			if err != nil {
				slog.Error("renew user service registration failed", "service", Service, "addr", grpcAddr, "error", err)
			}
			time.Sleep(time.Duration(CommonMicroConf.ServiceHub.HeartbeatFrequency)*time.Second - 200*time.Millisecond)
		}
	}()

	// 优雅退出
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(infraMySQL.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		AddFunc(func() {
			// 注销服务
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := ETCDServiceHub.Unregister(ctx, Service, grpcAddr); err != nil {
				slog.Error("unregister user service failed", "service", Service, "addr", grpcAddr, "error", err)
			}
		}).
		BuildBlock()
}
