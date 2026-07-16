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
	Service   = manager.InterviewService // 微服务名
	GoPostery = "go_postery"             // GoPostery 公共配置前缀
)

var (
	suffix       = ""
	ETCDEndpoint = hub.ETCDEndpoint // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production/evaluation")
	flag.Parse()

	ip, err := utils.GetLocalIP() // 获取本地内网 IP
	if err != nil {
		slog.Error("get local ip failed", "error", err)
		panic(err)
	}

	// 本地测试
	if *env != "production" {
		suffix = "_test"
		ip = "localhost"
		ETCDEndpoint = hub.LocalETCDEndpoint
	}
	ctx, cancel := context.WithCancel(context.Background())

	// RAG 离线评估
	if *env == "evaluation" {
		//evaluation.RunEval(ctx)
		return
	}

	// Remote Config Center
	etcdClient := infraEtcd.Init([]string{ETCDEndpoint}) // Init Etcd

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, etcdClient, GoPostery+suffix+"/")
	// 加载私有配置
	InterviewServiceConf := conf.LoadInterviewServiceConfig(ctx, etcdClient, Service+suffix+"/")

	infraSlog.InitSlog(InterviewServiceConf.Log) // Init Slog

	slog.Info("config loaded", "service", Service+suffix, "grpc_port", InterviewServiceConf.GRPC.Port, "metric_port", InterviewServiceConf.Metric.Port)
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service+suffix) // Init JaegerTracer

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
	MetricService := pkg.NewMetricService(Service + suffix)

	// gRPC Server
	InterviewServiceServer := server.NewInterviewServiceServer(InterviewService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+suffix+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	interview_grpc.RegisterInterviewServiceServer(ServiceRegistrar, InterviewServiceServer) // Register gRPC Service

	// Prometheus
	metricAddr := ip + ":" + InterviewServiceConf.Metric.Port
	slog.Info("metric server address resolved", "addr", metricAddr)
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("metric server failed", "addr", metricAddr, "error", err)
		}
	}()

	grpcAddr := ip + ":" + InterviewServiceConf.GRPC.Port
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
		slog.Error("register interview service failed", "service", Service, "addr", grpcAddr, "error", err)
		panic(err)
	}

	// 自动续约
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
