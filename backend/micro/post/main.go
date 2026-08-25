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
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
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
	"github.com/yzletter/go-postery/backend/micro/post/config"
	server "github.com/yzletter/go-postery/backend/micro/post/grpc"
	"github.com/yzletter/go-postery/backend/micro/post/repository"
	"github.com/yzletter/go-postery/backend/micro/post/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/post/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/post/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Service           = manager.PostService // 微服务名
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
	PostServiceConf := config.LoadPostServiceConfig(ctx, etcdClient, ServiceConfPrefix)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(PostServiceConf.Log)                                        // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service) // Init JaegerTracer

	// Infrastructure 层
	RedisClient := infraRedis.Init(CommonMicroConf.Redis) // Init Redis
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL) // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)   // 初始化 雪花算法

	// Cache 层
	PostCache := cache.NewPostCache(RedisClient)

	// DAO 层
	PostDAO := dao.NewPostDAO(MySQLGormDB)
	TagDAO := dao.NewTagDAO(MySQLGormDB)

	// Repository 层
	PostRepo := repository.NewPostRepository(PostDAO, PostCache, IDGenerator) // 注册 PostRepository
	TagRepo := repository.NewTagRepository(TagDAO)                            // 注册 TagRepository

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, etcdClient, hub.NewRoundRobinLoadBalancer())

	// gRPC Client 层
	InteractiveManager := manager.NewInteractiveManager(ctx, manager.InteractiveService, ETCDServiceHub)
	RankManager := manager.NewRankManager(ctx, manager.RankService, ETCDServiceHub)

	// Service 层
	PostService := service.NewPostService(PostRepo, TagRepo, IDGenerator, InteractiveManager, RankManager) // 注册 PostService
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 1000)
	MetricService := pkg.NewMetricService(Service) // 注册 MetricService

	// gRPC Server
	PostServiceServer := server.NewPostServiceServer(PostService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	post_grpc.RegisterPostServiceServer(ServiceRegistrar, PostServiceServer) // Register gRPC Service

	// Prometheus
	metricAddr := ip + ":" + PostServiceConf.Metric.Port
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("metric server failed", "addr", metricAddr, "error", err)
		}
	}()

	grpcAddr := ip + ":" + PostServiceConf.GRPC.Port
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
		slog.Error("register post service failed", "service", Service, "addr", grpcAddr, "error", err)
		panic(err)
	}

	// 向服务发现中心自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, Service, grpcAddr, leaseID)
			if err != nil {
				slog.Error("renew post service registration failed", "service", Service, "addr", grpcAddr, "error", err)
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
				slog.Error("unregister post service failed", "service", Service, "addr", grpcAddr, "error", err)
			}
		}).
		BuildBlock()
}
