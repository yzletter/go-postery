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
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/microservice-backend/code/conf"
	"github.com/yzletter/go-postery/microservice-backend/code/grpc"
	"github.com/yzletter/go-postery/microservice-backend/code/grpc/hub"
	"github.com/yzletter/go-postery/microservice-backend/code/infra/email"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/code/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/code/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/code/infra/jaeger"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/code/infra/redis"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/code/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/code/infra/sms"
	"github.com/yzletter/go-postery/microservice-backend/code/repository"
	"github.com/yzletter/go-postery/microservice-backend/code/repository/cache"
	"github.com/yzletter/go-postery/microservice-backend/code/service"
	"github.com/yzletter/go-postery/microservice-backend/code/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	ServiceName  string = "code_service" // 微服务名
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
	EtcdClient := infraEtcd.Init([]string{EtcdEndPoint})                                           // Init Etcd
	Config := conf.LoadGlobalConfig(ctx, EtcdClient, prefix+ServiceName+"_", prefix+GoPostery+"_") // Get Config From Remote Config Center
	fmt.Printf("%s Init Config Success %+v\n", prefix+ServiceName, Config)

	// Infra
	infraSlog.InitSlog(Config.Log)                                                   // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, prefix+ServiceName) // Init JaegerTracer
	RedisClient := infraRedis.Init(Config.Redis)                                     // Init Redis
	SmsClient := sms.NewAliyunSmsClient(Config.SMS)                                  // Init SMS
	EmailClient := email.NewSMTPEmailClient(Config.Email)                            // Init Email

	// Cache
	CodeCache := cache.NewCodeCache(RedisClient)
	// Repository
	CodeRepository := repository.NewCodeRepository(CodeCache)

	// Service
	CodeService := service.NewCodeService(CodeRepository, EmailClient, SmsClient)
	RateLimitService := service.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := service.NewMetricService(prefix + ServiceName)

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(Config.ServiceHub, EtcdClient, hub.NewRoundRobinLoadBalancer())
	ServiceHubProxy := hub.GetServiceHubProxy(ETCDServiceHub)

	// gRPC Server
	CodeServiceServer := grpc_server.NewCodeServiceServer(CodeService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(prefix+ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	code_grpc.RegisterCodeServiceServer(server, CodeServiceServer) // Register gRPC Service

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
	ip, err := utils.GetLocalIP() // 获取本地内网 IP
	if err != nil {
		slog.Error("Get Local IP Failed", "error", err)
		panic(err)
	}
	grpcAddr := ip + ":" + Config.GRPC.Port

	// 监听
	if lis, err := net.Listen("tcp", grpcAddr); err != nil {
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
	if leaseID, err := ServiceHubProxy.Register(ctx, ServiceName, grpcAddr, 0); err != nil {
		slog.Error("Service Code Server Register Failed", "service", ServiceName, "error", err)
		panic(err)
	} else {
		// 自动续约
		go func() {
			for {
				leaseID, err = ServiceHubProxy.Register(ctx, ServiceName, grpcAddr, leaseID)
				if err != nil {
					slog.Error("Service Code Server Register Failed", "service", ServiceName, "error", err)
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
