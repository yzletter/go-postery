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
	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	"github.com/yzletter/go-postery/microservice-backend/agent/config"
	grpc_server "github.com/yzletter/go-postery/microservice-backend/agent/grpc"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/agent/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/agent/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/agent/infra/jaeger"
	infraKafka "github.com/yzletter/go-postery/microservice-backend/agent/infra/kafka"
	llm2 "github.com/yzletter/go-postery/microservice-backend/agent/infra/llm"
	infraMySQL "github.com/yzletter/go-postery/microservice-backend/agent/infra/mysql"
	infraQdarant "github.com/yzletter/go-postery/microservice-backend/agent/infra/qdrant"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/agent/infra/redis"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/agent/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/agent/infra/snowflake"
	"github.com/yzletter/go-postery/microservice-backend/agent/repository"
	"github.com/yzletter/go-postery/microservice-backend/agent/repository/dao"
	service2 "github.com/yzletter/go-postery/microservice-backend/agent/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	ServiceName = "agent_service"
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

	// Infrastructure 层
	RedisClient := infraRedis.Init(Config.Redis)        // 初始化 Redis
	MySQLGormDB := infraMySQL.Init(Config.MySQL)        // 初始化 MySQL
	QdrantClient := infraQdarant.Init(Config.Qdrant)    // 初始化 Qdrant
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0) // 初始化 雪花算法

	QdrantKafkaConsumer := infraKafka.InitQdrantConsumer(Config.Kafka)
	AgentKafkaConsumer := infraKafka.InitAgentConsumer(Config.Kafka)

	ArkEmbedder := llm2.NewArkEmbedder(ctx, Config.Ark)  // 初始化火山引擎 Embedding 模型
	ArkChatModel := llm2.NewArkLLMModel(ctx, Config.Ark) // 初始化火山引擎 LLM 模型

	// DAO 层
	AgentDAO := dao.NewAgentDAO(ctx, MySQLGormDB, QdrantClient, ArkEmbedder.GetInternal())
	// Repository 层
	AgentRepo := repository.NewAgentRepository(AgentDAO)
	// Service 层
	AgentService := service2.NewAgentService(AgentRepo, AgentKafkaConsumer, QdrantKafkaConsumer, ArkEmbedder, ArkChatModel, IDGenerator)
	RateLimitService := service2.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := service2.NewMetricService(ServiceName)

	go AgentService.StartChunkDocConsumer(ctx)     // 开启切分文档协程
	go AgentService.StartUpsertQdrantConsumer(ctx) // 开启向量数据库协程

	// gRPC Server
	AgentServiceServer := grpc_server.NewAgentServiceServer(AgentService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	agent_grpc.RegisterAgentServiceServer(server, AgentServiceServer) // Register gRPC Service

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
