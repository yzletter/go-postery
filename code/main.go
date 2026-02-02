package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/code/config"
	"github.com/yzletter/go-postery/code/grpc"
	"github.com/yzletter/go-postery/code/infra/email"
	infraEtcd "github.com/yzletter/go-postery/code/infra/etcd"
	"github.com/yzletter/go-postery/code/infra/graceful_stop"
	infraRedis "github.com/yzletter/go-postery/code/infra/redis"
	infraSlog "github.com/yzletter/go-postery/code/infra/slog"
	"github.com/yzletter/go-postery/code/infra/sms"
	"github.com/yzletter/go-postery/code/repository"
	"github.com/yzletter/go-postery/code/repository/cache"
	"github.com/yzletter/go-postery/code/service"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{"172.16.150.246:2379"}) // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient)

	// Infra
	infraSlog.InitSlog(Config.Log)                        // Init Slog
	RedisClient := infraRedis.Init(Config.Redis)          // Init Redis
	SmsClient := sms.NewAliyunSmsClient(Config.SMS)       // Init SMS
	EmailClient := email.NewSMTPEmailClient(Config.Email) // Init Email
	// Cache
	CodeCache := cache.NewCodeCache(RedisClient)
	// Repository
	CodeRepository := repository.NewCodeRepository(CodeCache)
	// Service
	CodeService := service.NewCodeService(CodeRepository, EmailClient, SmsClient)
	MetricService := service.NewMetricService()
	// gRPC Server
	CodeServiceServer := grpc_server.NewCodeServiceServer(CodeService)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(MetricService.CounterInterceptor(), MetricService.TimerInterceptor()),
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
		AddFunc(infraRedis.Close).AddFunc(cancel).
		Build()

	// Start gRPC Server
	lis, err := net.Listen("tcp", Config.GRPC.Addr)
	if err != nil {
		panic(err)
	}
	if err := server.Serve(lis); err != nil {
		slog.Error("Code grpc Server Start Failed", "error", err)
		panic(err)
	}
}
