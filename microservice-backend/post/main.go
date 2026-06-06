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
	"github.com/yzletter/go-postery/microservice-backend/post/grpc/hub"
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
	ServiceName  string = "post_service" // 微服务名
	GoPostery    string = "go_postery"   // GoPostery 公共配置前缀
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
	MetricService := service.NewMetricService(prefix + ServiceName) // 注册 MetricService

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(Config.ServiceHub, EtcdClient, hub.NewRoundRobinLoadBalancer())
	ServiceHubProxy := hub.GetServiceHubProxy(ETCDServiceHub)

	// gRPC Server
	PostServiceServer := grpc_server.NewPostServiceServer(PostService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(prefix+ServiceName+":", RateLimitService).BuildLimiter),
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
		slog.Error("Service Post Server Register Failed", "service", ServiceName, "error", err)
		panic(err)
	} else {
		// 自动续约
		go func() {
			for {
				leaseID, err = ServiceHubProxy.Register(ctx, ServiceName, Config.GRPC.Addr, leaseID)
				if err != nil {
					slog.Error("Service Post Server Register Failed", "service", ServiceName, "error", err)
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
