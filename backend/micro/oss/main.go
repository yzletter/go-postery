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
	oss_grpc "github.com/yzletter/go-postery/api/proto/oss/v1"
	"github.com/yzletter/go-postery/backend/conf"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraOSS "github.com/yzletter/go-postery/backend/infra/oss"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/micro/oss/config"
	server "github.com/yzletter/go-postery/backend/micro/oss/grpc"
	"github.com/yzletter/go-postery/backend/micro/oss/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	Service           = manager.OSSService // 微服务名
	GoPostery         = conf.GoPostery     // GoPostery
	CommonConfPrefix  = GoPostery + "/conf/common_conf/"
	ServiceConfPrefix = GoPostery + "/conf/service_conf/" + Service + "_conf/"
)

func main() {
	// 启动参数, etcd 地址
	etcdEndpoint := flag.String("etcd", "localhost:2379", "etcd 地址")
	flag.Parse()

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
	OSSServiceConf := config.LoadOSSServiceConfig(ctx, etcdClient, ServiceConfPrefix)

	infraSlog.InitSlog(OSSServiceConf.Log)
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, Service)

	// Infrastructure 层
	RedisClient := infraRedis.Init(CommonMicroConf.Redis)
	OSSManager := infraOSS.Init(OSSServiceConf.OSS)

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, etcdClient, hub.NewRoundRobinLoadBalancer())

	// Service 层
	OSSService := service.NewOSSService(OSSManager, OSSServiceConf.OSS)
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 1000)
	MetricService := pkg.NewMetricService(Service)

	// gRPC Server
	OSSServiceServer := server.NewOSSServiceServer(OSSService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(Service+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	oss_grpc.RegisterOSSServiceServer(ServiceRegistrar, OSSServiceServer)

	// Prometheus
	metricAddr := ip + ":" + OSSServiceConf.Metric.Port
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("metric server failed", "addr", metricAddr, "error", err)
		}
	}()

	grpcAddr := ip + ":" + OSSServiceConf.GRPC.Port
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
		slog.Error("register oss service failed", "service", Service, "addr", grpcAddr, "error", err)
		panic(err)
	}

	// 向服务发现中心自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, Service, grpcAddr, leaseID)
			if err != nil {
				slog.Error("renew oss service registration failed", "service", Service, "addr", grpcAddr, "error", err)
			}
			time.Sleep(time.Duration(CommonMicroConf.ServiceHub.HeartbeatFrequency)*time.Second - 200*time.Millisecond)
		}
	}()

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		AddFunc(func() {
			// 注销服务
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := ETCDServiceHub.Unregister(ctx, Service, grpcAddr); err != nil {
				slog.Error("unregister oss service failed", "service", Service, "addr", grpcAddr, "error", err)
			}
		}).
		BuildBlock()
}
