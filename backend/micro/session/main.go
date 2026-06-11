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
	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"github.com/yzletter/go-postery/backend/conf"
	my_grpc "github.com/yzletter/go-postery/backend/grpc"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	infraMySQL "github.com/yzletter/go-postery/backend/infra/db/mysql"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraKafka "github.com/yzletter/go-postery/backend/infra/mq/kafka"
	infraRabbitMQ "github.com/yzletter/go-postery/backend/infra/mq/rabbitmq"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/infra/snowflake"
	server "github.com/yzletter/go-postery/backend/micro/session/grpc"
	"github.com/yzletter/go-postery/backend/micro/session/repository"
	"github.com/yzletter/go-postery/backend/micro/session/repository/dao"
	"github.com/yzletter/go-postery/backend/micro/session/service"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	ServiceName  = "session_service" // 微服务名
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
	EtcdClient := infraEtcd.Init([]string{EtcdEndPoint}) // Init Etcd

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, EtcdClient, prefix+GoPostery+"_")
	fmt.Printf("%s Init Common Config Success %+v\n", prefix+ServiceName, CommonMicroConf)
	// 加载私有配置
	SessionServiceConf := conf.LoadSessionServiceConfig(ctx, EtcdClient, prefix+ServiceName+"_")
	fmt.Printf("%s Init SessionService Config Success %+v\n", prefix+ServiceName, SessionServiceConf)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(SessionServiceConf.Log)                                                // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, prefix+ServiceName) // Init JaegerTracer

	// Infrastructure 层
	RedisClient := infraRedis.Init(CommonMicroConf.Redis)    // 初始化 Redis
	RabbitMQ := infraRabbitMQ.Init(CommonMicroConf.RabbitMQ) // 初始化 RabbitMQ
	MySQLGormDB := infraMySQL.Init(CommonMicroConf.MySQL)    // 初始化 MySQL
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)      // 初始化 雪花算法
	SessionKafkaConsumer := infraKafka.InitConsumer(CommonMicroConf.Kafka, conf.SessionKafkaTopic, conf.SessionKafkaGroup)

	// DAO 层
	MessageDAO := dao.NewMessageDAO(MySQLGormDB)
	SessionDAO := dao.NewSessionDAO(MySQLGormDB)
	// Repository 层
	MessageRepo := repository.NewMessageRepository(MessageDAO) // 注册 MessageRepository
	SessionRepo := repository.NewSessionRepository(SessionDAO) // 注册 SessionRepository
	// Service 层
	SessionService := service.NewSessionService(SessionRepo, MessageRepo, RabbitMQ, SessionKafkaConsumer, IDGenerator) // 注册 SessionService
	RateLimitService := ratelimit.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := pkg.NewMetricService(prefix + ServiceName)

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, EtcdClient, hub.NewRoundRobinLoadBalancer())

	go SessionService.StartSessionRegisterConsumer(ctx) // 开启协程注册新用户聊天功能

	// gRPC Server
	SessionServiceServer := server.NewSessionServiceServer(SessionService)
	ServiceRegistrar := grpc.NewServer(
		grpc.UnaryInterceptor(my_grpc.NewGrpcLimitInterceptor(prefix+ServiceName+":", RateLimitService).BuildLimiter),
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()), // Prometheus
		grpc.StatsHandler(otelgrpc.NewServerHandler()),                                                   // Jaeger
	)
	session_grpc.RegisterSessionServiceServer(ServiceRegistrar, SessionServiceServer) // Register gRPC Service

	// Prometheus
	metricAddr := ip + ":" + SessionServiceConf.Metric.Port
	slog.Info("Metric Addr Get Success", "addr", metricAddr)
	go func() {
		mux := http.NewServeMux()
		// Metric
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
		if err := http.ListenAndServe(metricAddr, mux); err != nil {
			slog.Error("Metric Server Failed", "error", err)
		}
	}()

	grpcAddr := ip + ":" + SessionServiceConf.GRPC.Port
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
		slog.Error("Service Session Server Register Failed", "service", ServiceName, "error", err)
		panic(err)
	}

	// 自动续约
	go func() {
		for {
			leaseID, err = ETCDServiceHub.Register(ctx, ServiceName, grpcAddr, leaseID)
			if err != nil {
				slog.Error("Service Session Server Register Failed", "service", ServiceName, "error", err)
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
				slog.Error("Service Session Server Unregister Failed", "service", ServiceName, "error", err)
			}
		}).
		BuildBlock()
}
