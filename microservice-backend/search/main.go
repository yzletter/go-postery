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
	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	"github.com/yzletter/go-postery/microservice-backend/search/config"
	grpc_server "github.com/yzletter/go-postery/microservice-backend/search/grpc"
	"github.com/yzletter/go-postery/microservice-backend/search/grpc/client"
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

const (
	ServiceName = "search_service"
	GoPostery   = "go_postery_"
	//GoPostery    = "test_go_postery_"
	EtcdEndPoint = "172.16.131.223:2379"
	//EtcdEndPoint = "localhost:12379"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{EtcdEndPoint})                           // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient, ServiceName+"_", GoPostery) // Get Config From Remote Config Center
	fmt.Printf("%s Init Config Success %+v\n", ServiceName, Config)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(Config.Log)                                            // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, ServiceName) // Init JaegerTracer

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
	MetricService := service2.NewMetricService()

	// gRPC Server
	SearchServiceServer := grpc_server.NewSearchServiceServer(SearchService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(ServiceName+":", RateLimitService).BuildLimiter),
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

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(cancel).AddFunc(TracerShutdown).
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
