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
	"github.com/yzletter/go-postery/backend/micro/rank/config"
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
	Service           = manager.RankService // 微服务名
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
	RankServiceConf := config.LoadRankServiceConfig(ctx, etcdClient, ServiceConfPrefix)

	// gRPC Common Infrastructures
	infraSlog.InitSlog(RankServiceConf.Log)                                        // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service) // Init JaegerTracer

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
	UserManager := manager.NewUserManager(ctx, manager.UserService, ETCDServiceHub)
	PostManager := manager.NewPostManager(ctx, manager.PostService, ETCDServiceHub)
	InteractiveManager := manager.NewInteractiveManager(ctx, manager.InteractiveService, ETCDServiceHub)

	// Service 层
	RankService := service.NewRankService(RankRepo, UserManager, PostManager, InteractiveManager, KafkaConsumer) // 注册 RankService
	go RankService.StartKafkaConsumer(ctx)                                                                       // 开启协程消费消息更新排行榜

	// 定时更新榜单
	Cron.AddFuncWithSpec("*/10 * * * *", RankService.CronRankTopK).Build()

	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 1000)
	MetricService := pkg.NewMetricService(Service) // 注册 MetricService

	// gRPC Server
	RankServiceServer := server.NewRankServiceServer(RankService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	rank_grpc.RegisterRankServiceServer(ServiceRegistrar, RankServiceServer) // Register gRPC Service

	// Prometheus
	metricAddr := ip + ":" + RankServiceConf.Metric.Port
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("Metric Server Failed", "error", err)
		}
	}()

	grpcAddr := ip + ":" + RankServiceConf.GRPC.Port
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
		slog.Error("Service Rank Server Register Failed", "service", Service, "error", err)
		panic(err)
	}

	// 向服务发现中心自动续约
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
