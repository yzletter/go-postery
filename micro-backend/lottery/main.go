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
	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/micro-backend/lottery/conf"
	hub2 "github.com/yzletter/go-postery/micro-backend/lottery/grpc/hub"
	"github.com/yzletter/go-postery/micro-backend/lottery/grpc/server"
	infraEtcd "github.com/yzletter/go-postery/micro-backend/lottery/infra/etcd"
	"github.com/yzletter/go-postery/micro-backend/lottery/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/micro-backend/lottery/infra/jaeger"
	infraMySQL "github.com/yzletter/go-postery/micro-backend/lottery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/micro-backend/lottery/infra/redis"
	infraRocketMQ "github.com/yzletter/go-postery/micro-backend/lottery/infra/rocketmq"
	infraSlog "github.com/yzletter/go-postery/micro-backend/lottery/infra/slog"
	"github.com/yzletter/go-postery/micro-backend/lottery/infra/snowflake"
	repository2 "github.com/yzletter/go-postery/micro-backend/lottery/repository"
	cache2 "github.com/yzletter/go-postery/micro-backend/lottery/repository/cache"
	dao2 "github.com/yzletter/go-postery/micro-backend/lottery/repository/dao"
	service2 "github.com/yzletter/go-postery/micro-backend/lottery/service"
	"github.com/yzletter/go-postery/micro-backend/lottery/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	ServiceName  = "lottery_service" // 微服务名
	GoPostery    = "go_postery"      // GoPostery 公共配置前缀
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
	GiftCache := cache2.NewGiftCache(RedisClient)
	OrderCache := cache2.NewOrderCache(RedisClient)
	// DAO 层
	GiftDAO := dao2.NewGiftDAO(MySQLGormDB)
	OrderDAO := dao2.NewOrderDAO(MySQLGormDB)
	// Repository 层
	GiftRepo := repository2.NewGiftRepository(GiftDAO, GiftCache)
	OrderRepo := repository2.NewOrderRepository(OrderDAO, OrderCache)

	// Service 层
	MetricService := service2.NewMetricService(prefix + ServiceName)
	RateLimitService := service2.NewRateLimitService(RedisClient, 5*time.Second, 5*5000)
	LotteryService := service2.NewLotteryService(OrderRepo, GiftRepo, RocketMQ, IDGenerator) // 注册 LotteryService
	//LotteryService.InitCacheInventory(ctx)
	go LotteryService.StartLotteryOrderConsumer(ctx) // 开启协程核查临时订单进行库存回流
	go LotteryService.StartStockRollbackScanner(ctx) // 开启协程兜底扫描失败库存回补

	// gRPC ServiceHub
	ETCDServiceHub := hub2.NewEtcdServiceHub(Config.ServiceHub, EtcdClient, hub2.NewRoundRobinLoadBalancer())

	// gRPC Server
	LotteryServiceServer := server.NewLotteryServiceServer(LotteryService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(server.NewGrpcLimitInterceptor(prefix+ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	lottery_grpc.RegisterLotteryServiceServer(ServiceRegistrar, LotteryServiceServer) // 注册服务

	// Prometheus
	metricAddr := ip + ":" + Config.Metric.Port
	slog.Info("Metric Addr Get Success", "addr", metricAddr)
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("Metric Server Failed", "error", err)
		}
	}()

	grpcAddr := ip + ":" + Config.GRPC.Port
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
		slog.Error("Service Lottery Server Register Failed", "service", ServiceName, "error", err)
		panic(err)
	}

	// 自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, ServiceName, grpcAddr, leaseID)
			if err != nil {
				slog.Error("Service Lottery Server Register Failed", "service", ServiceName, "error", err)
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
				slog.Error("Service Lottery Server Unregister Failed", "service", ServiceName, "error", err)
			}
		}).
		BuildBlock()
}
