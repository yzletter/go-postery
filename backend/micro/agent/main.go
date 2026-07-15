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
	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	"github.com/yzletter/go-postery/backend/conf"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/database/mysql"
	infraQdrant "github.com/yzletter/go-postery/backend/infra/database/qdrant"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraLLM "github.com/yzletter/go-postery/backend/infra/llm"
	infraKafka "github.com/yzletter/go-postery/backend/infra/mq/kafka"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/snowflake"
	server "github.com/yzletter/go-postery/backend/micro/agent/grpc"
	"github.com/yzletter/go-postery/backend/micro/agent/repository"
	"github.com/yzletter/go-postery/backend/micro/agent/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/agent/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Service   = manager.AgentService // 微服务名
	GoPostery = "go_postery"         // GoPostery 公共配置前缀
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

	// Remote Config Center
	etcdClient := infraEtcd.Init([]string{ETCDEndpoint}) // Init Etcd

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, etcdClient, GoPostery+suffix+"/")
	// 加载私有配置
	AgentServiceConf := conf.LoadAgentServiceConfig(ctx, etcdClient, Service+suffix+"/")

	// gRPC Common Infrastructure
	infraSlog.InitSlog(AgentServiceConf.Log) // Init Slog
	slog.Info("config loaded", "service", Service+suffix, "grpc_port", AgentServiceConf.GRPC.Port, "metric_port", AgentServiceConf.Metric.Port)
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service+suffix) // Init JaegerTracer

	// Infrastructure 层
	RedisClient := infraRedis.Init(CommonMicroConf.Redis)    // 初始化 Redis
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL)    // 初始化 MySQL
	QdrantClient := infraQdrant.Init(CommonMicroConf.Qdrant) // 初始化 Qdrant
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)      // 初始化 雪花算法
	QdrantKafkaConsumer := infraKafka.InitConsumer(CommonMicroConf.Kafka, conf.AgentQdrantKafkaTopic, conf.AgentQdrantKafkaGroup)
	AgentKafkaConsumer := infraKafka.InitConsumer(CommonMicroConf.Kafka, conf.AgentKafkaTopic, conf.AgentKafkaGroup)
	ArkEmbedder := infraLLM.NewArkEmbedder(ctx, AgentServiceConf.Ark)  // 初始化火山引擎 Embedding 模型
	ArkChatModel := infraLLM.NewArkLLMModel(ctx, AgentServiceConf.Ark) // 初始化火山引擎 LLM 模型

	// DAO 层
	AgentDAO := dao.NewAgentDAO(ctx, MySQLGormDB, QdrantClient, ArkEmbedder.GetInternal())
	// Repository 层
	AgentRepo := repository.NewAgentRepository(AgentDAO)

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, etcdClient, hub.NewRoundRobinLoadBalancer())

	// gRPC Client
	PostManager := manager.NewPostManager(ctx, manager.PostService, ETCDServiceHub)

	// Service 层
	AgentService := service.NewAgentService(AgentRepo, AgentKafkaConsumer, QdrantKafkaConsumer, ArkEmbedder, ArkChatModel, IDGenerator, PostManager)
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := pkg.NewMetricService(Service + suffix)

	go AgentService.StartChunkDocConsumer(ctx)     // 开启切分文档协程
	go AgentService.StartUpsertQdrantConsumer(ctx) // 开启向量数据库协程

	// gRPC Server
	AgentServiceServer := server.NewAgentServiceServer(AgentService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+suffix+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	agent_grpc.RegisterAgentServiceServer(ServiceRegistrar, AgentServiceServer) // Register gRPC Service

	// Prometheus
	metricAddr := ip + ":" + AgentServiceConf.Metric.Port
	slog.Info("metric server address resolved", "addr", metricAddr)
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("metric server failed", "addr", metricAddr, "error", err)
		}
	}()

	grpcAddr := ip + ":" + AgentServiceConf.GRPC.Port
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
		slog.Error("register agent service failed", "service", Service, "addr", grpcAddr, "error", err)
		panic(err)
	}

	// 自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, Service, grpcAddr, leaseID)
			if err != nil {
				slog.Error("renew agent service registration failed", "service", Service, "addr", grpcAddr, "error", err)
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
				slog.Error("unregister agent service failed", "service", Service, "addr", grpcAddr, "error", err)
			}
		}).
		BuildBlock()
}
