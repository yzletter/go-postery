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
	interview_grpc "github.com/yzletter/go-postery/api/proto/interview/v1"
	"github.com/yzletter/go-postery/backend/conf"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	infraMilvus "github.com/yzletter/go-postery/backend/infra/database/milvus"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/database/mysql"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraLLM "github.com/yzletter/go-postery/backend/infra/llm"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/snowflake"
	"github.com/yzletter/go-postery/backend/infra/tokenizer"
	"github.com/yzletter/go-postery/backend/micro/interview/config"
	"github.com/yzletter/go-postery/backend/micro/interview/dag"
	server "github.com/yzletter/go-postery/backend/micro/interview/grpc"
	"github.com/yzletter/go-postery/backend/micro/interview/loader"
	"github.com/yzletter/go-postery/backend/micro/interview/mcp"
	"github.com/yzletter/go-postery/backend/micro/interview/rag"
	"github.com/yzletter/go-postery/backend/micro/interview/repository"
	"github.com/yzletter/go-postery/backend/micro/interview/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/interview/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/interview/service"
	"github.com/yzletter/go-postery/backend/micro/interview/skill"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Service           = manager.InterviewService // 微服务名
	GoPostery         = conf.GoPostery           // GoPostery
	CommonConfPrefix  = GoPostery + "/conf/common_conf/"
	ServiceConfPrefix = GoPostery + "/conf/service_conf/" + Service + "_conf/"
)

func main() {
	// 启动参数, etcd 地址
	etcdEndpoint := flag.String("etcd", "localhost:2379", "etcd 地址")
	evaluation := flag.Bool("evaluation", false, "运行 RAG 离线评估")
	flag.Parse()

	// 获取本地内网 IP
	ip, err := utils.GetLocalIP()
	if err != nil {
		slog.Error("get local IP failed", "error", err)
		panic(err)
	}

	// 全局 Context
	ctx, cancel := context.WithCancel(context.Background())

	// RAG 离线评估
	if *evaluation {
		//evaluation.RunEval(ctx)
		return
	}

	// Init Etcd
	etcdClient := infraEtcd.Init([]string{*etcdEndpoint})

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, etcdClient, CommonConfPrefix)
	// 加载私有配置
	InterviewServiceConf := config.LoadInterviewServiceConfig(ctx, etcdClient, ServiceConfPrefix)

	infraSlog.InitSlog(InterviewServiceConf.Log) // Init Slog

	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service) // Init JaegerTracer

	// LLM
	QwenLLMModel := infraLLM.NewQwenLLMModel(ctx, InterviewServiceConf.Qwen) // 初始化千问大模型
	QwenEmbedder := infraLLM.NewQwenEmbedder(ctx, InterviewServiceConf.Qwen) // 初始化千问 Embedder
	MilvusClient := infraMilvus.NewMilvusClient(ctx, CommonMicroConf.Milvus)

	// Infrastructure 层
	RedisClient := infraRedis.Init(CommonMicroConf.Redis) // Init Redis
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL) // Init MySQL
	Tokenizer := tokenizer.NewSegoTokenizer()             // Init Tokenizer
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)   // 初始化雪花算法

	// Retriever
	MilvusRetriever := rag.NewMilvusRAGStore(MilvusClient, QwenEmbedder, 10) // MilvusRetriever
	BM25Manager := rag.NewBM25Manager(Tokenizer, 20)                         // BM25Retriever

	// Reranker
	LLMReranker := rag.NewLLMReranker(QwenLLMModel, 20)

	// MCP
	GithubSearcher, err := mcp.NewGitHubSearcher(InterviewServiceConf.Github.Token)
	if err != nil {
		slog.Warn("init github searcher failed", "error", err)
	}

	// Skill
	skillRegistry := skill.NewSkillRegistry()
	skillRegistry.Register(skill.NewQuickQuizSkill(QwenLLMModel, MilvusRetriever, BM25Manager))
	skillRegistry.Register(skill.NewConceptTutorSkill(QwenLLMModel, MilvusRetriever, BM25Manager))
	skillRegistry.Register(skill.NewProjectHighlightSkill(QwenLLMModel))
	skillRegistry.Register(skill.NewTechCompareSkill(QwenLLMModel, MilvusRetriever, BM25Manager))

	// Repository 层
	InterviewDAO := dao.NewInterviewDAO(MySQLGormDB)
	InterviewCache := cache.NewInterviewCache(RedisClient)
	InterviewRepository := repository.NewInterviewRepository(InterviewDAO, InterviewCache)

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, etcdClient, hub.NewRoundRobinLoadBalancer())
	WSGatewayManager := manager.NewWSGatewayManager(ctx, manager.WSGatewayService, ETCDServiceHub)
	OSSManager := manager.NewOSSManager(ctx, manager.OSSService, ETCDServiceHub)

	// DAG
	InterviewOrchestrator := dag.NewOrchestrator(QwenLLMModel, InterviewRepository, MilvusRetriever, BM25Manager, LLMReranker, IDGenerator)
	if GithubSearcher != nil {
		InterviewOrchestrator.ReviewPlannerAgent.SetGitHubSearcher(GithubSearcher)
	}

	InterviewOrchestrator.Callbacks = service.NewInterviewCallback(WSGatewayManager)

	PrepareGraph, err := InterviewOrchestrator.CompilePrepareGraph(ctx)
	if err != nil {
		slog.Error("compile prepare graph failed", "error", err)
		panic(err)
	}
	InterviewGraph, err := InterviewOrchestrator.CompileInterViewGraph(ctx)
	if err != nil {
		slog.Error("compile interview graph failed", "error", err)
		panic(err)
	}
	EvaluationGraph, err := InterviewOrchestrator.CompileEvaluationGraph(ctx)
	if err != nil {
		slog.Error("compile evaluation graph failed", "error", err)
		panic(err)
	}

	// Service 层
	QuestionParser := loader.NewQuestionParser(IDGenerator)
	InterviewService := service.NewInterviewService(WSGatewayManager, InterviewOrchestrator, skillRegistry, InterviewRepository, QwenLLMModel, QuestionParser, OSSManager, IDGenerator, PrepareGraph, InterviewGraph, EvaluationGraph)
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 1000)
	MetricService := pkg.NewMetricService(Service)

	// gRPC Server
	InterviewServiceServer := server.NewInterviewServiceServer(InterviewService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	interview_grpc.RegisterInterviewServiceServer(ServiceRegistrar, InterviewServiceServer) // Register gRPC Service

	// Prometheus
	metricAddr := ip + ":" + InterviewServiceConf.Metric.Port
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("metric server failed", "addr", metricAddr, "error", err)
		}
	}()

	grpcAddr := ip + ":" + InterviewServiceConf.GRPC.Port
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
		slog.Error("register interview service failed", "service", Service, "addr", grpcAddr, "error", err)
		panic(err)
	}

	// 向服务发现中心自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, Service, grpcAddr, leaseID)
			if err != nil {
				slog.Error("renew interview service registration failed", "service", Service, "addr", grpcAddr, "error", err)
			}
			time.Sleep(time.Duration(CommonMicroConf.ServiceHub.HeartbeatFrequency)*time.Second - 200*time.Millisecond)
		}
	}()

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(infraMySQL.Close).AddFunc(infraMilvus.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		AddFunc(func() {
			// 注销服务
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := ETCDServiceHub.Unregister(ctx, Service, grpcAddr); err != nil {
				slog.Error("unregister interview service failed", "service", Service, "addr", grpcAddr, "error", err)
			}
		}).
		BuildBlock()
}
