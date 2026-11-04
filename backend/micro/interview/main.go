package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yzletter/go-postery/backend/api/proto/interview/v1"
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
	Service       = manager.InterviewService // 微服务名
	ServicePrefix = Service + "/"            // 微服务配置前缀
)

func runEval(args []string) {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	prepare := fs.Bool("prepare", false, "准备评估环境：解析 MD 题库写入 Milvus/BM25，导出 manifest.json")
	genDataset := fs.Bool("gen-dataset", false, "基于 manifest 自动生成评估数据集（需先运行 --prepare）")
	datasetPath := fs.String("dataset", "data/eval/dataset_v1.json", "评估数据集路径")
	outDir := fs.String("out", "data/eval/reports", "报告输出目录")
	note := fs.String("note", "", "实验备注，写入报告 Config.Note，便于 A/B 对比")
	skipRerank := fs.Bool("skip-rerank", false, "跳过 LLM Rerank 阶段（用于 A/B 对比）")
	questionsDir := fs.String("questions", "data/questions", "题库 MD 文件目录（仅 --prepare 模式使用）")
	manifestPath := fs.String("manifest", "data/eval/manifest.json", "manifest 输出/读取路径")
	sampleCount := fs.Int("sample-count", 50, "自动生成数据集的样本数量（仅 --gen-dataset 使用）")
	if err := fs.Parse(args); err != nil {
		log.Fatalf("解析参数失败: %v", err)
	}

	if *prepare {
		runEvalPrepare(*questionsDir, *manifestPath)
	} else if *genDataset {
		runEvalGenDataset(*manifestPath, *datasetPath, *sampleCount)
	} else {
		runEvalExecute(*datasetPath, *outDir, *note, *skipRerank, *manifestPath)
	}
}

func main() {
	// RAG 评估
	if len(os.Args) > 1 && os.Args[1] == "eval" {
		runEval(os.Args[2:])
		return
	}

	// 启动时指定远程配置中心地址
	etcdEndpoint := flag.String("etcd", "localhost:2379", "远程配置中心地址，默认为本地")
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
	CommonMicroConf := conf.LoadCommonConf(ctx, etcdClient, conf.CommonConfPrefix)
	// 加载私有配置
	InterviewServiceConf := config.LoadInterviewServiceConfig(ctx, etcdClient, conf.ServiceConfPrefix+ServicePrefix)

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
