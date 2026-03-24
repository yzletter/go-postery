package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/microservice-backend/code/config"
	"github.com/yzletter/go-postery/microservice-backend/code/grpc"
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
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const ServiceName = "code_service"

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{"172.16.131.223:2379"})       // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient, ServiceName+"_") // Get Config From Remote Config Center
	fmt.Printf("%s Init Config Success %+v\n", ServiceName, Config)

	// Infra
	infraSlog.InitSlog(Config.Log)                                            // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, ServiceName) // Init JaegerTracer
	RedisClient := infraRedis.Init(Config.Redis)                              // Init Redis
	SmsClient := sms.NewAliyunSmsClient(Config.SMS)                           // Init SMS
	EmailClient := email.NewSMTPEmailClient(Config.Email)                     // Init Email

	// Cache
	CodeCache := cache.NewCodeCache(RedisClient)

	// Repository
	CodeRepository := repository.NewCodeRepository(CodeCache)

	// Service
	CodeService := service.NewCodeService(CodeRepository, EmailClient, SmsClient)
	RateLimitService := service.NewRateLimitService(RedisClient, time.Minute, 10)
	MetricService := service.NewMetricService(ServiceName)

	// gRPC Server
	CodeServiceServer := grpc_server.NewCodeServiceServer(CodeService)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_server.NewGrpcLimitInterceptor(ServiceName+":", RateLimitService).BuildLimiter),
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

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(cancel).AddFunc(TracerShutdown).
		Build()

	// Start gRPC Server
	lis, err := net.Listen("tcp", Config.GRPC.Addr)
	if err != nil {
		panic(err)
	}
	if err := server.Serve(lis); err != nil {
		slog.Error("Service gRPC Server Start Failed", "service", ServiceName, "error", err)
		panic(err)
	}
}
