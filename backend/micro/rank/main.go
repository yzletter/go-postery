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
	rank_grpc "github.com/yzletter/go-postery/api/proto/rank/v1"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/event"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	"github.com/yzletter/go-postery/backend/infra/crontab"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraKafka "github.com/yzletter/go-postery/backend/infra/mq/kafka"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	server "github.com/yzletter/go-postery/backend/micro/rank/grpc"
	"github.com/yzletter/go-postery/backend/micro/rank/repository"
	"github.com/yzletter/go-postery/backend/micro/rank/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/rank/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Service   = manager.RankService // 微服务名
	GoPostery = "go_postery"        // GoPostery 公共配置前缀
)

var (
	suffix       = ""
	ETCDEndpoint = hub.ETCDEndpoint // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production")
	flag.Parse()

	ip, err := utils.GetLocalIP() // 获取本地内网 IP
	if err != nil {
		slog.Error("get local ip failed", "error", err)
		panic(err)
	}

	// 本地测试
	if *env == "local" {
		suffix = "_test"
		ip = "localhost"
		ETCDEndpoint = "localhost:12379"
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 初始化远程配置中心
	etcdClient := infraEtcd.Init([]string{ETCDEndpoint})

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, etcdClient, GoPostery+suffix+"/")

	// 加载私有配置
	RankServiceConf := conf.LoadRankServiceConfig(ctx, etcdClient, Service+suffix+"/")

	// gRPC Common Infrastructures
	infraSlog.InitSlog(RankServiceConf.Log) // Init Slog
	slog.Info("config loaded", "service", Service+suffix, "grpc_port", RankServiceConf.GRPC.Port, "metric_port", RankServiceConf.Metric.Port)
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service+suffix) // Init JaegerTracer

	// Infrastructure 层
	RedisClient := infraRedis.Init(CommonMicroConf.Redis) // Init Redis
	KafkaConsumer := infraKafka.InitConsumer(CommonMicroConf.Kafka, event.KafkaTopicRankUpdateScore, event.KafkaRankGroup)
	Cron := crontab.NewCrontabBuilder() // init Crontab

	// Cache 层
	RankCache := cache.NewRankCache(RedisClient)

	// Repository 层
	RankRepo := repository.NewRankRepository(RankCache) // 注册 RankRepository

	// ServiceHub 层
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, etcdClient, hub.NewRoundRobinLoadBalancer())

	// gRPC Client
	ETCDServiceHub.LoadEndpoints(ctx, manager.UserService)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, manager.UserService)
	UserManager := manager.NewUserManager(manager.UserService, ETCDServiceHub)
	go UserManager.StartHealthCheck(ctx) // 开启下游服务健康检查

	ETCDServiceHub.LoadEndpoints(ctx, manager.PostService)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, manager.PostService)
	PostManager := manager.NewPostManager(manager.PostService, ETCDServiceHub)
	go PostManager.StartHealthCheck(ctx) // 开启下游服务健康检查

	ETCDServiceHub.LoadEndpoints(ctx, manager.InteractiveService)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, manager.InteractiveService)
	InteractiveManager := manager.NewInteractiveManager(manager.InteractiveService, ETCDServiceHub)
	go InteractiveManager.StartHealthCheck(ctx) // 开启下游服务健康检查

	// Service 层
	RankService := service.NewRankService(RankRepo, UserManager, PostManager, InteractiveManager, KafkaConsumer) // 注册 RankService
	go RankService.StartKafkaConsumer(ctx)                                                                       // 开启协程消费消息更新排行榜

	// 定时更新榜单
	Cron.AddFuncWithSpec("*/10 * * * *", RankService.CronRankTopK).Build()

	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := pkg.NewMetricService(Service + suffix) // 注册 MetricService

	// gRPC Server
	RankServiceServer := server.NewRankServiceServer(RankService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+suffix+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	rank_grpc.RegisterRankServiceServer(ServiceRegistrar, RankServiceServer) // Register gRPC Service

	// Prometheus
	metricAddr := ip + ":" + RankServiceConf.Metric.Port
	slog.Info("Metric Addr Get Success", "addr", metricAddr)
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("Metric Server Failed", "error", err)
		}
	}()

	grpcAddr := ip + ":" + RankServiceConf.GRPC.Port
	slog.Info("gRPC Addr Get Success", "addr", grpcAddr)
	if lis, err := net.Listen("tcp", grpcAddr); err != nil {
		panic(err)
	} else {
		go func() {
			if err := ServiceRegistrar.Serve(lis); err != nil {
				slog.Error("Service gRPC Server Start Failed", "service", Service+suffix, "error", err)
				panic(err)
			}
		}()
	}

	// 向服务中心注册服务, 这里不加环境后缀
	leaseID, err := ETCDServiceHub.Register(ctx, Service, grpcAddr, 0)
	if err != nil {
		slog.Error("Service Rank Server Register Failed", "service", Service, "error", err)
		panic(err)
	}

	// 自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, Service, grpcAddr, leaseID)
			if err != nil {
				slog.Error("Service Rank Server Register Failed", "service", Service, "error", err)
			}
			time.Sleep(time.Duration(CommonMicroConf.ServiceHub.HeartbeatFrequency)*time.Second - 200*time.Millisecond)
		}
	}()

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(infraKafka.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		AddFunc(func() {
			// 注销服务
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := ETCDServiceHub.Unregister(ctx, Service, grpcAddr); err != nil {
				slog.Error("Service Rank Server Unregister Failed", "service", Service, "error", err)
			}
		}).
		AddFunc(Cron.Stop).
		BuildBlock()
}
