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
	"github.com/yzletter/go-postery/backend/conf"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/db/mysql"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/snowflake"
	server "github.com/yzletter/go-postery/backend/micro/post/grpc"
	"github.com/yzletter/go-postery/backend/micro/post/repository"
	"github.com/yzletter/go-postery/backend/micro/post/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/post/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/post/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	ServiceName  = "post_service" // 微服务名
	GoPostery    = "go_postery"   // GoPostery 公共配置前缀
	prefix       = ""
	EtcdEndPoint string // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production")
	flag.Parse()

	ip, err := utils.GetLocalIP() // 获取本地内网 IP
	if err != nil {
		slog.Error("Get Local IP Failed", "error", err)
		panic(err)
	}

	// 本地测试
	if *env == "local" {
		prefix = "test_"
		ip = "localhost"
		EtcdEndPoint = "localhost:12379"
	} else {
		EtcdEndPoint = "172.16.131.223:2379"
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{EtcdEndPoint}) // Init Etcd

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, EtcdClient, prefix+GoPostery+"_")
	fmt.Printf("%s Init Common Config Success %+v\n", prefix+ServiceName, CommonMicroConf)
	// 加载私有配置
	PostServiceConf := conf.LoadPostServiceConfig(ctx, EtcdClient, prefix+ServiceName+"_")
	fmt.Printf("%s Init PostService Config Success %+v\n", prefix+ServiceName, PostServiceConf)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(PostServiceConf.Log)                                                   // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, prefix+ServiceName) // Init JaegerTracer

	// Infrastructure 层
	RedisClient := infraRedis.Init(CommonMicroConf.Redis) // Init Redis
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL) // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)   // 初始化 雪花算法

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
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := pkg.NewMetricService(prefix + ServiceName) // 注册 MetricService

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, EtcdClient, hub.NewRoundRobinLoadBalancer())

	// gRPC Server
	PostServiceServer := server.NewPostServiceServer(PostService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(prefix+ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	post_grpc.RegisterPostServiceServer(ServiceRegistrar, PostServiceServer) // Register gRPC Service

	// Prometheus
	metricAddr := ip + ":" + PostServiceConf.Metric.Port
	slog.Info("Metric Addr Get Success", "addr", metricAddr)
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("Metric Server Failed", "error", err)
		}
	}()

	grpcAddr := ip + ":" + PostServiceConf.GRPC.Port
	slog.Info("gRPC Addr Get Success", "addr", grpcAddr)
	if lis, err := net.Listen("tcp", grpcAddr); err != nil {
		panic(err)
	} else {
		go func() {
			if err := ServiceRegistrar.Serve(lis); err != nil {
				slog.Error("Service gRPC Server Start Failed", "service", prefix+ServiceName, "error", err)
				panic(err)
			}
		}()
	}

	// 向服务中心注册服务, 这里不加前缀 prefix
	leaseID, err := ETCDServiceHub.Register(ctx, ServiceName, grpcAddr, 0)
	if err != nil {
		slog.Error("Service Post Server Register Failed", "service", ServiceName, "error", err)
		panic(err)
	}

	// 自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, ServiceName, grpcAddr, leaseID)
			if err != nil {
				slog.Error("Service Post Server Register Failed", "service", ServiceName, "error", err)
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
			if err := ETCDServiceHub.Unregister(ctx, ServiceName, grpcAddr); err != nil {
				slog.Error("Service Post Server Unregister Failed", "service", ServiceName, "error", err)
			}
		}).
		BuildBlock()
}
