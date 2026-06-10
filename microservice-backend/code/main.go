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
	"github.com/yzletter/go-postery/microservice-backend/code/grpc/hub"
	"github.com/yzletter/go-postery/microservice-backend/code/grpc/server"
	"github.com/yzletter/go-postery/microservice-backend/code/infra/email"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/code/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/code/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/code/infra/jaeger"
	infraMySQL "github.com/yzletter/go-postery/microservice-backend/code/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/code/infra/redis"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/code/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/code/infra/sms"
	"github.com/yzletter/go-postery/microservice-backend/code/repository"
	"github.com/yzletter/go-postery/microservice-backend/code/repository/cache"
	"github.com/yzletter/go-postery/microservice-backend/code/repository/dao"
	"github.com/yzletter/go-postery/microservice-backend/code/service"
	"github.com/yzletter/go-postery/microservice-backend/code/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	ServiceName  = "code_service" // 微服务名
	GoPostery    = "go_postery"   // GoPostery 公共配置前缀
	prefix       = ""
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

	// Infrastructure
	infraSlog.InitSlog(Config.Log)                                                   // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, prefix+ServiceName) // Init JaegerTracer
	RedisClient := infraRedis.Init(Config.Redis)                                     // Init Redis
	MySQLGormDB := infraMySQL.Init(Config.MySQL)                                     // Init MySQL
	SmsClient := sms.NewAliyunSmsClient(Config.SMS)                                  // Init SMS
	EmailClient := email.NewSMTPEmailClient(Config.Email)                            // Init Email

	// Cache
	CodeCache := cache.NewCodeCache(RedisClient)
	// DAO
	CodeDAO := dao.NewCodeDAO(MySQLGormDB)
	// Repository
	CodeRepository := repository.NewCodeRepository(CodeDAO, CodeCache)
	// Service
	CodeService := service.NewCodeService(CodeRepository, EmailClient, SmsClient)
	// Common Service
	RateLimitService := service.NewRateLimitService(RedisClient, time.Minute, 50)
	MetricService := service.NewMetricService(prefix + ServiceName)

	// gRPC ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(Config.ServiceHub, EtcdClient, hub.NewRoundRobinLoadBalancer())
	// gRPC Server
	CodeServiceServer := server.NewCodeServiceServer(CodeService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(server.NewGrpcLimitInterceptor(prefix+ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	code_grpc.RegisterCodeServiceServer(ServiceRegistrar, CodeServiceServer) // Register gRPC Service

	// Start gRPC Server
	ip, err := utils.GetLocalIP() // 获取本地内网 IP
	if *env == "local" {
		ip = "localhost"
	}
	fmt.Println(ip)
	if err != nil {
		slog.Error("Get Local IP Failed", "error", err)
		panic(err)
	}

	// Prometheus
	metricAddr := ip + ":" + Config.Metric.Port
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("Metric Server Failed", "error", err)
		}
	}()

	// 监听 gRPC
	grpcAddr := ip + ":" + Config.GRPC.Port
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
		slog.Error("Service Code Server Register Failed", "service", ServiceName, "error", err)
		panic(err)
	}

	// 自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, ServiceName, grpcAddr, leaseID)
			if err != nil {
				slog.Error("Service Code Server Register Failed", "service", ServiceName, "error", err)
			}
			time.Sleep(time.Duration(Config.ServiceHub.HeartbeatFrequency)*time.Second - 200*time.Millisecond)
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
				slog.Error("Service Code Server Unregister Failed", "service", ServiceName, "error", err)
			}
		}).
		BuildBlock()
}
