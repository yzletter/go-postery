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
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/backend/conf"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/database/mysql"
	"github.com/yzletter/go-postery/backend/infra/email"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/sms"
	"github.com/yzletter/go-postery/backend/infra/snowflake"
	server "github.com/yzletter/go-postery/backend/micro/code/grpc"
	"github.com/yzletter/go-postery/backend/micro/code/repository"
	"github.com/yzletter/go-postery/backend/micro/code/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/code/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/code/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Service   = manager.CodeService // 微服务名
	GoPostery = "go_postery"        // GoPostery 公共配置前缀
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
		slog.Error("get local IP failed", "error", err)
		panic(err)
	}

	// 本地测试
	if *env == "local" {
		suffix = "_test"
		ip = "localhost"
		ETCDEndpoint = "localhost:12379"
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 初始化 etcd
	etcdClient := infraEtcd.Init([]string{ETCDEndpoint})

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, etcdClient, GoPostery+suffix+"/")
	// 加载私有配置
	CodeServiceConf := conf.LoadCodeServiceConfig(ctx, etcdClient, Service+suffix+"/")

	// Infrastructure
	infraSlog.InitSlog(CodeServiceConf.Log) // Init Slog
	slog.Info("config loaded", "service", Service+suffix, "grpc_port", CodeServiceConf.GRPC.Port, "metric_port", CodeServiceConf.Metric.Port)
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service+suffix) // Init JaegerTracer
	RedisClient := infraRedis.Init(CommonMicroConf.Redis)                                 // Init Redis
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL)                                 // Init MySQL
	SmsClient := sms.NewAliyunSmsClient(CodeServiceConf.SMS)                              // Init SMS
	EmailClient := email.NewSMTPEmailClient(CodeServiceConf.Email)                        // Init Email
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)                                   // 初始化 雪花算法

	// Cache
	CodeCache := cache.NewCodeCache(RedisClient)
	// DAO
	CodeDAO := dao.NewCodeDAO(MySQLGormDB)
	// Repository
	CodeRepository := repository.NewCodeRepository(CodeDAO, CodeCache)
	// Service
	CodeService := service.NewCodeService(CodeRepository, EmailClient, SmsClient, IDGenerator)
	// Common Service
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 50)
	MetricService := pkg.NewMetricService(Service + suffix)

	// gRPC ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, etcdClient, hub.NewRoundRobinLoadBalancer())
	// gRPC Server
	CodeServiceServer := server.NewCodeServiceServer(CodeService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+suffix+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	code_grpc.RegisterCodeServiceServer(ServiceRegistrar, CodeServiceServer) // Register gRPC Service

	// Prometheus
	metricAddr := ip + ":" + CodeServiceConf.Metric.Port
	slog.Info("metric address get success", "addr", metricAddr)
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("metric server failed", "error", err)
		}
	}()

	// 监听 gRPC
	grpcAddr := ip + ":" + CodeServiceConf.GRPC.Port
	slog.Info("gRPC address get success", "addr", grpcAddr)
	if lis, err := net.Listen("tcp", grpcAddr); err != nil {
		panic(err)
	} else {
		go func() {
			if err := ServiceRegistrar.Serve(lis); err != nil {
				slog.Error("service gRPC server start failed", "service", Service+suffix, "error", err)
				panic(err)
			}
		}()
	}

	// 向服务中心注册服务, 这里不加环境后缀
	leaseID, err := ETCDServiceHub.Register(ctx, Service, grpcAddr, 0)
	if err != nil {
		slog.Error("service gRPC server register failed", "service", Service, "error", err)
		panic(err)
	}

	// 自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, Service, grpcAddr, leaseID)
			if err != nil {
				slog.Error("service gRPC server relet failed", "service", Service, "error", err)
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
				slog.Error("service gRPC server unregister failed", "service", Service, "error", err)
			}
		}).
		BuildBlock()
}
