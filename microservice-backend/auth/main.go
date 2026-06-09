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
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	"github.com/yzletter/go-postery/microservice-backend/auth/conf"
	grpc_server "github.com/yzletter/go-postery/microservice-backend/auth/grpc"
	"github.com/yzletter/go-postery/microservice-backend/auth/grpc/client"
	"github.com/yzletter/go-postery/microservice-backend/auth/grpc/hub"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/auth/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/auth/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/auth/infra/jaeger"
	infraMySQL "github.com/yzletter/go-postery/microservice-backend/auth/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/auth/infra/redis"
	"github.com/yzletter/go-postery/microservice-backend/auth/infra/security"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/auth/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/auth/infra/snowflake"
	"github.com/yzletter/go-postery/microservice-backend/auth/repository"
	"github.com/yzletter/go-postery/microservice-backend/auth/repository/cache"
	"github.com/yzletter/go-postery/microservice-backend/auth/repository/dao"
	"github.com/yzletter/go-postery/microservice-backend/auth/service"
	"github.com/yzletter/go-postery/microservice-backend/auth/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	ServiceName  string = "auth_service" // 微服务名
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

	// gRPC Common Infrastructure
	infraSlog.InitSlog(Config.Log)                                                   // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, prefix+ServiceName) // Init JaegerTracer

	// Infrastructure 层
	RedisClient := infraRedis.Init(Config.Redis)           // Init Redis
	MySQLGormDB := infraMySQL.Init(Config.MySQL)           // Init MySQL
	JwtManager := security.NewJwtManager(conf.JwtTokenKey) // Init JWT
	PasswordHasher := security.NewBcryptPasswordHasher(0)  // Init PasswordHasher
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)    // Init Snowflake

	// Cache 层
	AuthCache := cache.NewAuthCache(RedisClient)
	// DAO 层
	AuthDAO := dao.NewAuthDAO(MySQLGormDB)
	// Repository 层
	AuthRepo := repository.NewAuthRepository(AuthDAO, AuthCache)

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(Config.ServiceHub, EtcdClient, hub.NewRoundRobinLoadBalancer())
	ServiceHubProxy := hub.GetServiceHubProxy(ETCDServiceHub)

	// gRPC Client
	ConnCenter := client.NewConnectionCenter(ServiceHubProxy)
	CodeConn, err := ConnCenter.NewConnection(ctx, client.CodeServiceName)
	if err != nil {
		slog.Error("Init Code gRPC Connection Failed", "error", err)
	}
	CodeClient, err := client.NewCodeClient(CodeConn)
	if err != nil {
		slog.Error("Init Code gRPC Client Failed", "error", err)
	}

	// Service 层
	AuthService := service.NewAuthService(AuthRepo, JwtManager, PasswordHasher, IDGenerator, CodeClient) // 注册 AuthService
	RateLimitService := service.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := service.NewMetricService(prefix + ServiceName)

	// gRPC Server
	AuthServiceServer := grpc_server.NewAuthServiceServer(AuthService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(prefix+ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	auth_grpc.RegisterAuthServiceServer(server, AuthServiceServer) // 注册服务

	// Start gRPC Server
	ip, err := utils.GetLocalIP() // 获取本地内网 IP
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
		slog.Error("Service Auth Server Register Failed", "service", ServiceName, "error", err)
		panic(err)
	} else {
		// 自动续约
		go func() {
			for {
				leaseID, err = ServiceHubProxy.Register(ctx, ServiceName, grpcAddr, leaseID)
				if err != nil {
					slog.Error("Service Auth Server Register Failed", "service", ServiceName, "error", err)
				}
				time.Sleep(time.Duration(Config.ServiceHub.HeartbeatFrequency)*time.Second - 200*time.Millisecond)
			}
		}()
	}

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(infraMySQL.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		BuildBlock()
}
