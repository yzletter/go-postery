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
	interactive_grpc "github.com/yzletter/go-postery/api/proto/interactive/v1"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/event"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/database/mysql"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraKafka "github.com/yzletter/go-postery/backend/infra/mq/kafka"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/snowflake"
	server "github.com/yzletter/go-postery/backend/micro/interactive/grpc"
	"github.com/yzletter/go-postery/backend/micro/interactive/repository"
	"github.com/yzletter/go-postery/backend/micro/interactive/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/interactive/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/interactive/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Service   = manager.InteractiveService // 微服务名
	GoPostery = "go_postery"               // GoPostery 公共配置前缀
)

var (
	suffix       = ""
	ETCDEndpoint = hub.ETCDEndpoint // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production")
	flag.Parse()

	// 获取本地内网 IP
	ip, err := utils.GetLocalIP()
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

	// Remote Config Center
	etcdClient := infraEtcd.Init([]string{ETCDEndpoint})

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, etcdClient, GoPostery+suffix+"/")

	// 加载私有配置
	InteractiveServiceConf := conf.LoadInteractiveServiceConfig(ctx, etcdClient, Service+suffix+"/")

	// gRPC Common Infrastructure
	infraSlog.InitSlog(InteractiveServiceConf.Log)
	slog.Info("config loaded", "service", Service+suffix, "grpc_port", InteractiveServiceConf.GRPC.Port, "metric_port", InteractiveServiceConf.Metric.Port)
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service+suffix)

	// Infrastructure 层
	RedisClient := infraRedis.Init(CommonMicroConf.Redis)
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL)
	KafkaConsumer := infraKafka.InitConsumer(CommonMicroConf.Kafka,
		event.KafkaTopicInteractiveLike,                                                                        // topic
		event.KafkaInteractiveGroup,                                                                            // group
		event.KafkaTopicInteractiveFollow, event.KafkaTopicInteractiveComment, event.KafkaTopicInteractiveRead, // topic
	)
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)

	// Cache 层
	InteractiveCache := cache.NewInteractiveCache(RedisClient)

	// DAO 层
	InteractiveDAO := dao.NewInteractiveDAO(MySQLGormDB, IDGenerator)

	// Repository 层
	InteractiveRepo := repository.NewInteractiveRepository(InteractiveDAO, InteractiveCache)

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, etcdClient, hub.NewRoundRobinLoadBalancer())

	// gRPC Client
	ETCDServiceHub.LoadEndpoints(ctx, manager.PostService)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, manager.PostService)
	PostManager := manager.NewPostManager(manager.PostService, ETCDServiceHub)
	go PostManager.StartHealthCheck(ctx) // 开启下游服务健康检查

	// Service 层
	InteractiveService := service.NewInteractiveService(InteractiveRepo, PostManager, IDGenerator, KafkaConsumer)
	go InteractiveService.StartKafkaConsumer(ctx) // 开启协程消费消息更新互动数

	// Common Service
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := pkg.NewMetricService(Service + suffix)

	// gRPC Server
	InteractiveServiceServer := server.NewInteractiveServiceServer(InteractiveService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+suffix+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	interactive_grpc.RegisterInteractiveServiceServer(ServiceRegistrar, InteractiveServiceServer) // Register gRPC Service

	// Prometheus
	metricAddr := ip + ":" + InteractiveServiceConf.Metric.Port
	slog.Info("metric server address resolved", "addr", metricAddr)
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
			promhttp.Handler().ServeHTTP(w, r)
		})
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("metric server failed", "addr", metricAddr, "error", err)
		}
	}()

	// 启动 gRPC 服务
	grpcAddr := ip + ":" + InteractiveServiceConf.GRPC.Port
	slog.Info("grpc server address resolved", "addr", grpcAddr)
	if lis, err := net.Listen("tcp", grpcAddr); err != nil {
		panic(err)
	} else {
		go func() {
			if err := ServiceRegistrar.Serve(lis); err != nil {
				slog.Error("grpc server failed", "service", Service+suffix, "addr", grpcAddr, "error", err)
				panic(err)
			}
		}()
	}

	// 向服务中心注册服务, 这里不加环境后缀
	leaseID, err := ETCDServiceHub.Register(ctx, Service, grpcAddr, 0)
	if err != nil {
		slog.Error("register interactive service failed", "service", Service, "addr", grpcAddr, "error", err)
		panic(err)
	}

	// 自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, Service, grpcAddr, leaseID)
			if err != nil {
				slog.Error("renew interactive service registration failed", "service", Service, "addr", grpcAddr, "error", err)
			}
			time.Sleep(time.Duration(CommonMicroConf.ServiceHub.HeartbeatFrequency)*time.Second - 200*time.Millisecond)
		}
	}()

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(infraMySQL.Close).AddFunc(infraKafka.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		AddFunc(func() {
			// 注销服务
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := ETCDServiceHub.Unregister(ctx, Service, grpcAddr); err != nil {
				slog.Error("unregister interactive service failed", "service", Service, "addr", grpcAddr, "error", err)
			}
		}).
		BuildBlock()
}
