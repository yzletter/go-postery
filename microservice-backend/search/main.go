package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	"github.com/yzletter/go-postery/microservice-backend/search/config"
	grpc_server "github.com/yzletter/go-postery/microservice-backend/search/grpc"
	"github.com/yzletter/go-postery/microservice-backend/search/grpc/client"
	"github.com/yzletter/go-postery/microservice-backend/search/grpc/hub"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/search/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/search/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/search/infra/jaeger"
	"github.com/yzletter/go-postery/microservice-backend/search/infra/kafka"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/search/infra/redis"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/search/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/search/infra/snowflake"
	"github.com/yzletter/go-postery/microservice-backend/search/infra/tokenizer"
	service2 "github.com/yzletter/go-postery/microservice-backend/search/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	ServiceName  string = "search_service" // 微服务名
	GoPostery    string = "go_postery"     // GoPostery 公共配置前缀
	prefix       string = ""
	EtcdEndPoint string // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production")
	flag.Parse()

	// 本地测试
	if *env == "local" {
		prefix = "test_"
		EtcdEndPoint = "localhost:12379"
	} else {
		EtcdEndPoint = "172.16.131.223:2379"
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{EtcdEndPoint})                                             // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient, prefix+ServiceName+"_", prefix+GoPostery+"_") // Get Config From Remote Config Center
	fmt.Printf("%s Init Config Success %+v\n", prefix+ServiceName, Config)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(Config.Log)                                                   // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, prefix+ServiceName) // Init JaegerTracer

	// Infrastructure
	RedisClient := infraRedis.Init(Config.Redis) // 初始化 Redis
	KafkaConsumer := kafka.InitConsumer(Config.Kafka)
	Tokenizer := tokenizer.NewJiebaTokenizer()          // 初始化分词器
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0) // 初始化 雪花算法

	// gRPC Client
	PostClient, err := client.NewPostClient()
	if err != nil {
		slog.Error("Init Post gRPC Client Failed", "error", err)
	}

	// Service 层
	SearchService := service2.NewSearchService(KafkaConsumer, Tokenizer, IDGenerator, PostClient)
	go SearchService.StartConsumer(ctx) // 开启协程消费消息对新文章进行索引

	RateLimitService := service2.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := service2.NewMetricService(prefix + ServiceName)

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(Config.ServiceHub, EtcdClient, hub.NewRoundRobinLoadBalancer())
	ServiceHubProxy := hub.GetServiceHubProxy(ETCDServiceHub)

	// gRPC Server
	SearchServiceServer := grpc_server.NewSearchServiceServer(SearchService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(prefix+ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	search_grpc.RegisterSearchServiceServer(server, SearchServiceServer) // Register gRPC Service

	// Prometheus
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(Config.Metric.Addr, mux); err != nil {
			slog.Error("Metric Server Failed", "error", err)
		}
	}()

	// Start gRPC Server
	if lis, err := net.Listen("tcp", Config.GRPC.Addr); err != nil {
		panic(err)
	} else {
		go func() {
			if err := server.Serve(lis); err != nil {
				slog.Error("Service gRPC Server Start Failed", "service", prefix+ServiceName, "error", err)
				panic(err)
			}
		}()
	}

	// 向服务中心注册服务, 这里不加前缀 prefix
	if leaseID, err := ServiceHubProxy.Register(ctx, ServiceName, Config.GRPC.Addr, 0); err != nil {
		slog.Error("Service Search Server Register Failed", "service", ServiceName, "error", err)
		panic(err)
	} else {
		// 自动续约
		go func() {
			for {
				leaseID, err = ServiceHubProxy.Register(ctx, ServiceName, Config.GRPC.Addr, leaseID)
				if err != nil {
					slog.Error("Service Search Server Register Failed", "service", ServiceName, "error", err)
				}
				time.Sleep(time.Duration(Config.ServiceHub.HeartbeatFrequency)*time.Second - 200*time.Millisecond)
			}
		}()
	}

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		BuildBlock()
}
