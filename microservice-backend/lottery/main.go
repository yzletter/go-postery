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
	auth_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/microservice-backend/lottery/conf"
	grpc_server "github.com/yzletter/go-postery/microservice-backend/lottery/grpc"
	"github.com/yzletter/go-postery/microservice-backend/lottery/grpc/hub"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/lottery/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/lottery/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/lottery/infra/jaeger"
	infraMySQL "github.com/yzletter/go-postery/microservice-backend/lottery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/lottery/infra/redis"
	infraRocketMQ "github.com/yzletter/go-postery/microservice-backend/lottery/infra/rocketmq"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/lottery/infra/slog"
	"github.com/yzletter/go-postery/microservice-backend/lottery/infra/snowflake"
	"github.com/yzletter/go-postery/microservice-backend/lottery/repository"
	"github.com/yzletter/go-postery/microservice-backend/lottery/repository/cache"
	"github.com/yzletter/go-postery/microservice-backend/lottery/repository/dao"
	"github.com/yzletter/go-postery/microservice-backend/lottery/service"
	"github.com/yzletter/go-postery/microservice-backend/lottery/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	ServiceName  string = "lottery_service" // 微服务名
	GoPostery    string = "go_postery"      // GoPostery 公共配置前缀
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
	RedisClient := infraRedis.Init(Config.Redis)        // Init Redis
	MySQLGormDB := infraMySQL.Init(Config.MySQL)        // Init MySQL
	RocketMQ := infraRocketMQ.Init(Config.RocketMQ)     // Init RocketMQ
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0) // Init Snowflake

	// Cache 层
	GiftCache := cache.NewGiftCache(RedisClient)
	OrderCache := cache.NewOrderCache(RedisClient)
	// DAO 层
	GiftDAO := dao.NewGiftDAO(MySQLGormDB)
	OrderDAO := dao.NewOrderDAO(MySQLGormDB)
	// Repository 层
	GiftRepo := repository.NewGiftRepository(GiftDAO, GiftCache)
	OrderRepo := repository.NewOrderRepository(OrderDAO, OrderCache)

	// Service 层
	MetricService := service.NewMetricService(prefix + ServiceName)
	RateLimitService := service.NewRateLimitService(RedisClient, 5*time.Second, 5*5000)
	LotteryService := service.NewLotteryService(OrderRepo, GiftRepo, RocketMQ, IDGenerator) // 注册 LotteryService
	LotteryService.InitCacheInventory(context.Background())
	go LotteryService.StartLotteryOrderConsumer(context.Background()) // 开启协程核查临时订单进行库存回流
	go LotteryService.StartStockRollbackScanner(context.Background()) // 开启协程兜底扫描失败库存回补

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(Config.ServiceHub, EtcdClient, hub.NewRoundRobinLoadBalancer())
	ServiceHubProxy := hub.GetServiceHubProxy(ETCDServiceHub)

	// gRPC Server
	LotteryServiceServer := grpc_server.NewLotteryServiceServer(LotteryService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(prefix+ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	auth_grpc.RegisterLotteryServiceServer(server, LotteryServiceServer) // 注册服务

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
		slog.Error("Service Lottery Server Register Failed", "service", ServiceName, "error", err)
		panic(err)
	} else {
		// 自动续约
		go func() {
			for {
				leaseID, err = ServiceHubProxy.Register(ctx, ServiceName, grpcAddr, leaseID)
				if err != nil {
					slog.Error("Service Lottery Server Register Failed", "service", ServiceName, "error", err)
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
