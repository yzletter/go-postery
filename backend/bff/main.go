package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ws_gateway_grpc "github.com/yzletter/go-postery/api/proto/ws_gateway/v1"
	"github.com/yzletter/go-postery/backend/bff/handler"
	"github.com/yzletter/go-postery/backend/bff/middleware"
	ws_gateway_grpc_server "github.com/yzletter/go-postery/backend/bff/ws_gateway/grpc"
	ws_gateway_handler "github.com/yzletter/go-postery/backend/bff/ws_gateway/handler"
	service2 "github.com/yzletter/go-postery/backend/bff/ws_gateway/service"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
	"github.com/yzletter/go-postery/backend/pkg"
	"github.com/yzletter/go-postery/backend/pkg/ratelimit"
	"github.com/yzletter/go-postery/backend/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	ServiceName  = manager.BFFService // 微服务名
	GoPostery    = "go_postery"       // GoPostery 公共配置前缀
	suffix       = ""
	ETCDEndpoint = hub.ETCDEndpoint // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production")
	flag.Parse()

	ip, err := utils.GetLocalIP()
	if err != nil {
		slog.Error("get local ip failed", "error", err)
		panic(err)
	}

	// 本地测试
	if *env == "local" {
		suffix = "_test"
		ip = "localhost"
		ETCDEndpoint = hub.LocalETCDEndpoint
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{ETCDEndpoint}) // Init Etcd

	// 加载公共配置
	CommonMicroConf := conf.LoadCommonMicroConf(ctx, EtcdClient, GoPostery+suffix+"/")
	fmt.Printf("%s Init Common Config Success %+v\n", ServiceName+suffix, CommonMicroConf)
	// 加载私有配置
	BFFServiceConf := conf.LoadBFFServiceConfig(ctx, EtcdClient, ServiceName+suffix+"/")
	fmt.Printf("%s Init BFFService Config Success %+v\n", ServiceName+suffix, BFFServiceConf)
	// 加载 Websocket 网关配置
	WSGatewayServiceConf := conf.LoadWSGatewayServiceConfig(ctx, EtcdClient, manager.WSGatewayService+suffix+"/")
	fmt.Printf("%s Init WSGatewayService Config Success %+v\n", manager.WSGatewayService+suffix, WSGatewayServiceConf)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(BFFServiceConf.Log)                                                    // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, ServiceName+suffix) // Init JaegerTracer
	RedisClient := infraRedis.Init(CommonMicroConf.Redis)                                     // 初始化 Redis

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, EtcdClient, hub.NewRoundRobinLoadBalancer())

	// gRPC Service 层
	AuthServiceClient := manager.NewAuthManager(ctx, manager.AuthService, ETCDServiceHub)
	CodeServiceClient := manager.NewCodeManager(ctx, manager.CodeService, ETCDServiceHub)
	UserServiceClient := manager.NewUserManager(ctx, manager.UserService, ETCDServiceHub)
	InteractiveServiceClient := manager.NewInteractiveManager(ctx, manager.InteractiveService, ETCDServiceHub)
	PostServiceClient := manager.NewPostManager(ctx, manager.PostService, ETCDServiceHub)
	SearchServiceClient := manager.NewSearchManager(ctx, manager.SearchService, ETCDServiceHub)
	InterviewServiceClient := manager.NewInterviewManager(ctx, manager.InterviewService, ETCDServiceHub)
	LotteryServiceClient := manager.NewLotteryManager(ctx, manager.LotteryService, ETCDServiceHub)
	SessionServiceClient := manager.NewSessionManager(ctx, manager.SessionService, ETCDServiceHub)

	// Service 层
	BFFMetricRegistry := prometheus.NewRegistry()
	BFFMetricRegistry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	MetricSvc := pkg.NewMetricServiceWithRegistry(ServiceName+suffix, BFFMetricRegistry)                   // 注册 MetricService
	RateLimitSvc := ratelimit.NewRateLimitService(RedisClient, conf.RateLimitInterval, conf.RateLimitRate) // 注册 RateLimitService

	// Handler 层
	PostHdl := handler.NewPostHandler(PostServiceClient, UserServiceClient, InteractiveServiceClient) // 注册 PostHandler
	AuthHdl := handler.NewAuthHandler(AuthServiceClient, CodeServiceClient, UserServiceClient)        // 注册 AuthHandler
	UserHdl := handler.NewUserHandler(UserServiceClient, PostServiceClient, InteractiveServiceClient) // 注册 UserHandler
	SearchHdl := handler.NewSearchHandler(SearchServiceClient, PostServiceClient, UserServiceClient)  // 注册 SearchHandler
	InterviewHdl := handler.NewInterviewHandler(InterviewServiceClient)                               // 注册 InterviewHandler
	LotteryHdl := handler.NewLotteryHandler(LotteryServiceClient, UserServiceClient)                  // 注册 LotteryHandler
	SessionHdl := handler.NewSessionHandler(SessionServiceClient, UserServiceClient)                  // 注册 SessionHandler

	// 中间件层
	AuthRequiredMdl := middleware.AuthRequiredMiddleware(AuthServiceClient) // AuthRequiredMdl 强制登录中间件
	MetricMdl := middleware.MetricMiddleware(MetricSvc)                     // MetricMdl 用于 Prometheus 监控中间件
	RateLimitMdl := middleware.RateLimitMiddleware(RateLimitSvc)            // RateLimitMdl 限流中间件
	CorsMdl := cors.New(cors.Config{                                        // CorsMdl 跨域中间件
		AllowOrigins:     []string{"http://" + BFFServiceConf.App.FrontendAddr}, // 允许域名跨域
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "traceparent", "tracestate", "baggage"},
		AllowCredentials: true, // 是否允许携带 cookie 之类的用户认证信息
		ExposeHeaders:    []string{"Content-Length", "Authorization", "traceparent", "tracestate", "baggage"},
		MaxAge:           6 * time.Hour,
	})

	// Websocket 网关
	WSGatewayHandler := ws_gateway_handler.NewHandler(InterviewServiceClient, SessionServiceClient)
	WSGateway := service2.NewWebsocketGateway(WSGatewayHandler)
	WSGatewayMetricRegistry := prometheus.NewRegistry()
	WSGatewayMetricSvc := pkg.NewMetricServiceWithRegistry(manager.WSGatewayService+suffix, WSGatewayMetricRegistry)
	WSGatewayServiceServer := ws_gateway_grpc_server.NewWSGatewayServiceServer(WSGateway)
	WSGatewayServiceRegistrar := grpc.NewServer(
		grpc.ChainUnaryInterceptor(WSGatewayMetricSvc.CounterInterceptor(), WSGatewayMetricSvc.TimerInterceptor()),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	ws_gateway_grpc.RegisterWSGatewayServiceServer(WSGatewayServiceRegistrar, WSGatewayServiceServer)

	WSGatewayMetricAddr := ip + ":" + WSGatewayServiceConf.Metric.Port
	WSGatewayMetricMux := http.NewServeMux()
	WSGatewayMetricMux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.HandlerFor(WSGatewayMetricRegistry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})
	WSGatewayMetricServer := &http.Server{
		Addr:    WSGatewayMetricAddr,
		Handler: WSGatewayMetricMux,
	}
	slog.Info("websocket metric server address resolved", "addr", WSGatewayMetricAddr)
	go func() {
		if err := WSGatewayMetricServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("websocket metric server failed", "addr", WSGatewayMetricAddr, "error", err)
		}
	}()

	// 服务端启动服务
	WSGatewayGRPCAddr := ip + ":" + WSGatewayServiceConf.GRPC.Port
	slog.Info("websocket grpc server address resolved", "addr", WSGatewayGRPCAddr)
	if lis, err := net.Listen("tcp", WSGatewayGRPCAddr); err != nil {
		panic(err)
	} else {
		go func() {
			if err := WSGatewayServiceRegistrar.Serve(lis); err != nil {
				slog.Error("websocket grpc server failed", "service", manager.WSGatewayService+suffix, "addr", WSGatewayGRPCAddr, "error", err)
				panic(err)
			}
		}()
	}

	// 服务端注册
	WSGatewayLeaseID, err := ETCDServiceHub.Register(ctx, manager.WSGatewayService, WSGatewayGRPCAddr, 0)
	if err != nil {
		slog.Error("register websocket gateway service failed", "service", manager.WSGatewayService, "addr", WSGatewayGRPCAddr, "error", err)
		panic(err)
	}

	// 服务端续约
	go func() {
		for {
			WSGatewayLeaseID, err = ETCDServiceHub.Register(ctx, manager.WSGatewayService, WSGatewayGRPCAddr, WSGatewayLeaseID)
			if err != nil {
				slog.Error("renew websocket gateway service registration failed", "service", manager.WSGatewayService, "addr", WSGatewayGRPCAddr, "error", err)
			}
			time.Sleep(time.Duration(CommonMicroConf.ServiceHub.HeartbeatFrequency)*time.Second - 200*time.Millisecond)
		}
	}()

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(TracerShutdown).AddFunc(cancel).
		AddFunc(func() {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer shutdownCancel()
			if err := ETCDServiceHub.Unregister(shutdownCtx, manager.WSGatewayService, WSGatewayGRPCAddr); err != nil {
				slog.Error("unregister websocket gateway service failed", "service", manager.WSGatewayService, "addr", WSGatewayGRPCAddr, "error", err)
			}
		}).
		AddFunc(WSGatewayServiceRegistrar.GracefulStop).
		Build()

	// 初始化 gin
	engine := gin.Default()
	engine.ContextWithFallback = true

	// 注册全局中间件
	engine.Use(
		middleware.TracingMiddleware(ServiceName+suffix), // OpenTelemetry tracing 中间件
		CorsMdl,      // CorsMdl 跨域中间件
		MetricMdl,    // Prometheus 监控中间件
		RateLimitMdl, // 限流中间件
	)

	// 运维接口
	engine.GET("/metrics", func(ctx *gin.Context) { // Prometheus 访问的接口
		promhttp.HandlerFor(BFFMetricRegistry, promhttp.HandlerOpts{}).ServeHTTP(ctx.Writer, ctx.Request)
	})

	// 业务接口
	api := engine.Group("/api")
	v1 := api.Group("/v1")

	UserHdl.RegisterRouter(v1, AuthRequiredMdl)      // 用户模块路由
	AuthHdl.RegisterRouter(v1, AuthRequiredMdl)      // 身份认证模块路由
	InterviewHdl.RegisterRouter(v1, AuthRequiredMdl) // 面试模块路由
	PostHdl.RegisterRouter(v1, AuthRequiredMdl)      // 帖子模块路由
	SearchHdl.RegisterRouter(v1, AuthRequiredMdl)    // 搜索模块路由
	LotteryHdl.RegisterRouter(v1, AuthRequiredMdl)   // 抽奖模块路由
	SessionHdl.RegisterRouter(v1, AuthRequiredMdl)   // 会话模块路由

	// 网关
	ws := v1.Group("")
	ws.Use(AuthRequiredMdl)
	ws.GET("/ws/session", WSGateway.NewSessionConnectionGin)
	ws.GET("/ws/interview", WSGateway.NewInterviewConnectionGin)

	if err := engine.Run(BFFServiceConf.App.BackendAddr); err != nil {
		panic(err)
	}
}
