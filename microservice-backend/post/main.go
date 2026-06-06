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
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"github.com/yzletter/go-postery/microservice-backend/post/config"
	grpc_server "github.com/yzletter/go-postery/microservice-backend/post/grpc"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/post/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/post/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/post/infra/jaeger"
	infraMySQL "github.com/yzletter/go-postery/microservice-backend/post/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/post/infra/redis"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/post/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/post/infra/snowflake"
	"github.com/yzletter/go-postery/microservice-backend/post/repository"
	"github.com/yzletter/go-postery/microservice-backend/post/repository/cache"
	"github.com/yzletter/go-postery/microservice-backend/post/repository/dao"
	"github.com/yzletter/go-postery/microservice-backend/post/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	ServiceName  string // 微服务名
	GoPostery    string // GoPostery 公共配置前缀
	EtcdEndPoint string // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production")
	flag.Parse()

	// 本地测试
	if *env == "local" {
		ServiceName = "test_post_service"
		GoPostery = "test_go_postery"
		EtcdEndPoint = "localhost:12379"
	} else {
		ServiceName = "post_service"
		GoPostery = "go_postery"
		EtcdEndPoint = "172.16.131.223:2379"
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{EtcdEndPoint})                               // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient, ServiceName+"_", GoPostery+"_") // Get Config From Remote Config Center
	fmt.Printf("%s Init Config Success %+v\n", ServiceName, Config)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(Config.Log)                                            // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, ServiceName) // Init JaegerTracer

	// Infrastructure 层
	RedisClient := infraRedis.Init(Config.Redis)        // Init Redis
	MySQLGormDB := infraMySQL.Init(Config.MySQL)        // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0) // 初始化 雪花算法

	// Cache 层
	PostCache := cache.NewPostCache(RedisClient)

	// DAO 层
	PostDAO := dao.NewPostDAO(MySQLGormDB)
	LikeDAO := dao.NewLikeDAO(MySQLGormDB)
	TagDAO := dao.NewTagDAO(MySQLGormDB)
	CommentDAO := dao.NewCommentDAO(MySQLGormDB)

	// Repository 层
	PostRepo := repository.NewPostRepository(PostDAO, PostCache) // 注册 PostRepository
	LikeRepo := repository.NewLikeRepository(LikeDAO)            // 注册 LikeRepository
	TagRepo := repository.NewTagRepository(TagDAO)               // 注册 TagRepository
	CommentRepo := repository.NewCommentRepository(CommentDAO)   // 注册 CommentRepository

	// Service 层
	PostService := service.NewPostService(PostRepo, LikeRepo, TagRepo, CommentRepo, IDGenerator) // 注册 PostService
	RateLimitService := service.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := service.NewMetricService(ServiceName) // 注册 MetricService

	// gRPC Server
	PostServiceServer := grpc_server.NewPostServiceServer(PostService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	post_grpc.RegisterPostServiceServer(server, PostServiceServer) // Register gRPC Service

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
