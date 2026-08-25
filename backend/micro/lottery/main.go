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
	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/backend/conf"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/database/mysql"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraRocketMQ "github.com/yzletter/go-postery/backend/infra/mq/rocketmq"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/snowflake"
	"github.com/yzletter/go-postery/backend/micro/lottery/config"
	server "github.com/yzletter/go-postery/backend/micro/lottery/grpc"
	"github.com/yzletter/go-postery/backend/micro/lottery/repository"
	"github.com/yzletter/go-postery/backend/micro/lottery/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/lottery/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/lottery/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Service           = manager.LotteryService // 微服务名
	GoPostery         = conf.GoPostery         // GoPostery
	CommonConfPrefix  = GoPostery + "/conf/common_conf/"
	ServiceConfPrefix = GoPostery + "/conf/service_conf/" + Service + "_conf/"
)

func main() {
	// 启动参数, etcd 地址
	etcdEndpoint := flag.String("etcd", "localhost:2379", "etcd 地址")
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
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, etcdClient, CommonConfPrefix)
	// 加载私有配置
	LotteryServiceConf := config.LoadLotteryServiceConfig(ctx, etcdClient, ServiceConfPrefix)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(LotteryServiceConf.Log)                                     // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service) // Init JaegerTracer

	// Infrastructure 层
	RedisClient := infraRedis.Init(CommonMicroConf.Redis) // Init Redis
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL) // Init MySQL
	RocketMQ := infraRocketMQ.Init(CommonMicroConf.RocketMQ,
		conf.RocketLotteryTopic, conf.RocketLotteryConsumerGroup, conf.RocketAwaitDuration) // Init RocketMQ
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
	MetricService := pkg.NewMetricService(Service)
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, 5*time.Second, 5*5000)
	LotteryService := service.NewLotteryService(OrderRepo, GiftRepo, RocketMQ, IDGenerator) // 注册 LotteryService
	//LotteryService.InitCacheInventory(ctx)
	go LotteryService.StartLotteryOrderConsumer(ctx) // 开启协程核查临时订单进行库存回流
	go LotteryService.StartStockRollbackScanner(ctx) // 开启协程兜底扫描失败库存回补

	// gRPC ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, etcdClient, hub.NewRoundRobinLoadBalancer())

	// gRPC Server
	LotteryServiceServer := server.NewLotteryServiceServer(LotteryService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	lottery_grpc.RegisterLotteryServiceServer(ServiceRegistrar, LotteryServiceServer) // 注册服务

	// Prometheus
	metricAddr := ip + ":" + LotteryServiceConf.Metric.Port
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("metric server failed", "addr", metricAddr, "error", err)
		}
	}()

	grpcAddr := ip + ":" + LotteryServiceConf.GRPC.Port
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
		slog.Error("register lottery service failed", "service", Service, "addr", grpcAddr, "error", err)
		panic(err)
	}

	// 向服务发现中心自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, Service, grpcAddr, leaseID)
			if err != nil {
				slog.Error("renew lottery service registration failed", "service", Service, "addr", grpcAddr, "error", err)
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
				slog.Error("unregister lottery service failed", "service", Service, "addr", grpcAddr, "error", err)
			}
		}).
		BuildBlock()
}
