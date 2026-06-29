package main

import (
	"context"
	"flag"
	"fmt"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yzletter/go-postery/backend/bff/handler"
	"github.com/yzletter/go-postery/backend/bff/middleware"
	"github.com/yzletter/go-postery/backend/bff/service"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraRedis "github.com/yzletter/go-postery/backend/infra/cache/redis"
	"github.com/yzletter/go-postery/backend/infra/crontab"
	"github.com/yzletter/go-postery/backend/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/backend/infra/jaeger"
	infraRabbitMQ "github.com/yzletter/go-postery/backend/infra/mq/rabbitmq"
	infraSlog "github.com/yzletter/go-postery/backend/infra/slog"
)

var (
	ServiceName  = "bff_service" // 微服务名
	GoPostery    = "go_postery"  // GoPostery 公共配置前缀
	suffix       = ""
	ETCDEndpoint = hub.ETCDEndpoint // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production")
	flag.Parse()

	// 本地测试
	if *env == "local" {
		suffix = "_test"
		ETCDEndpoint = "localhost:12379"
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

	// gRPC Common Infrastructure
	infraSlog.InitSlog(BFFServiceConf.Log)                                                    // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, CommonMicroConf.Jaeger, ServiceName+suffix) // Init JaegerTracer
	RedisClient := infraRedis.Init(CommonMicroConf.Redis)                                     // 初始化 Redis

	// Infrastructure 层
	RabbitMQ := infraRabbitMQ.Init(CommonMicroConf.RabbitMQ) // 初始化 RabbitMQ

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(CommonMicroConf.ServiceHub.HeartbeatFrequency, CommonMicroConf.ServiceHub.ServiceRegisterPrefix, EtcdClient, hub.NewRoundRobinLoadBalancer())

	// GRPC Service 层
	AuthServiceName := "auth_service"
	ETCDServiceHub.LoadEndpoints(ctx, AuthServiceName)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, AuthServiceName)
	AuthServiceClient := manager.NewAuthManager(AuthServiceName, ETCDServiceHub)
	go AuthServiceClient.StartHealthCheck(ctx) // 开启下游服务健康检查

	CodeServiceName := "code_service"
	ETCDServiceHub.LoadEndpoints(ctx, CodeServiceName)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, CodeServiceName)
	CodeServiceClient := manager.NewCodeManager(CodeServiceName, ETCDServiceHub)
	go CodeServiceClient.StartHealthCheck(ctx) // 开启下游服务健康检查

	UserServiceName := "user_service"
	ETCDServiceHub.LoadEndpoints(ctx, UserServiceName)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, UserServiceName)
	UserServiceClient := manager.NewUserManager(UserServiceName, ETCDServiceHub)
	go UserServiceClient.StartHealthCheck(ctx) // 开启下游服务健康检查

	InteractiveServiceName := "interactive_service"
	ETCDServiceHub.LoadEndpoints(ctx, InteractiveServiceName)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, InteractiveServiceName)
	InteractiveServiceClient := manager.NewInteractiveManager(InteractiveServiceName, ETCDServiceHub)
	go InteractiveServiceClient.StartHealthCheck(ctx) // 开启下游服务健康检查

	PostServiceName := "post_service"
	ETCDServiceHub.LoadEndpoints(ctx, PostServiceName)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, PostServiceName)
	PostServiceClient := manager.NewPostManager(PostServiceName, ETCDServiceHub)
	go PostServiceClient.StartHealthCheck(ctx) // 开启下游服务健康检查

	SearchServiceName := "search_service"
	ETCDServiceHub.LoadEndpoints(ctx, SearchServiceName)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, SearchServiceName)
	SearchServiceClient := manager.NewSearchManager(SearchServiceName, ETCDServiceHub)
	go SearchServiceClient.StartHealthCheck(ctx) // 开启下游服务健康检查

	AgentServiceName := "agent_service"
	ETCDServiceHub.LoadEndpoints(ctx, AgentServiceName)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, AgentServiceName)
	AgentServiceClient := manager.NewAgentManager(AgentServiceName, ETCDServiceHub)
	go AgentServiceClient.StartHealthCheck(ctx) // 开启下游服务健康检查

	LotteryServiceName := "lottery_service"
	ETCDServiceHub.LoadEndpoints(ctx, LotteryServiceName)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, LotteryServiceName)
	LotteryServiceClient := manager.NewLotteryManager(LotteryServiceName, ETCDServiceHub)
	go LotteryServiceClient.StartHealthCheck(ctx) // 开启下游服务健康检查

	SessionServiceName := "session_service"
	ETCDServiceHub.LoadEndpoints(ctx, SessionServiceName)
	ETCDServiceHub.WatchEndpointsFromServiceHub(ctx, SessionServiceName)
	SessionServiceClient := manager.NewSessionManager(SessionServiceName, ETCDServiceHub)
	go SessionServiceClient.StartHealthCheck(ctx) // 开启下游服务健康检查

	// Service 层
	MetricSvc := service.NewMetricService()                                                              // 注册 MetricService
	RateLimitSvc := service.NewRateLimitService(RedisClient, conf.RateLimitInterval, conf.RateLimitRate) // 注册 RateLimitService
	WebsocketSvc := service.NewWebsocketService(SessionServiceClient, RabbitMQ)                          // 注册 WebsocketService

	// Handler 层
	PostHdl := handler.NewPostHandler(PostServiceClient, UserServiceClient, InteractiveServiceClient) // 注册 PostHandler
	AuthHdl := handler.NewAuthHandler(AuthServiceClient, CodeServiceClient, UserServiceClient)        // 注册 AuthHandler
	UserHdl := handler.NewUserHandler(UserServiceClient, PostServiceClient, InteractiveServiceClient) // 注册 UserHandler
	SearchHdl := handler.NewSearchHandler(SearchServiceClient, PostServiceClient, UserServiceClient)  // 注册 SearchHandler
	AgentHdl := handler.NewAgentHandler(AgentServiceClient)                                           // 注册 AgentHandler
	LotteryHdl := handler.NewLotteryHandler(LotteryServiceClient, UserServiceClient)                  // 注册 LotteryHandler
	SessionHdl := handler.NewSessionHandler(SessionServiceClient, UserServiceClient)                  // 注册 SessionHandler
	WebsocketHdl := handler.NewWebsocketHandler(WebsocketSvc)                                         // 注册 WebsocketHandler

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

	// 初始化 Crontab
	crontab.NewCrontabBuilder().
		AddFuncWithSpec("*/10 * * * *", infraRedis.Ping).
		Build()

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRedis.Close).AddFunc(infraRabbitMQ.Close).AddFunc(cancel).AddFunc(TracerShutdown).
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
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request) // 固定写法
	})

	// 业务接口
	api := engine.Group("/api")
	v1 := api.Group("/v1")

	// 注册路由

	// 注册用户模块路由
	users := v1.Group("/users")
	// api/v1/users
	UserHdl.RegisterPublicRouter(users)
	// api/v1/users/authed
	authedUsers := users.Group("/authed")
	authedUsers.Use(AuthRequiredMdl)
	UserHdl.RegisterPrivateRouter(authedUsers)

	// 私信模块
	chat := users.Group("/:id/sessions")
	chat.Use(AuthRequiredMdl)
	{
		chat.GET("", SessionHdl.GetSession)                 // GET /api/v1/users/:id/sessions
		chat.GET("/messages", SessionHdl.GetHistoryMessage) // GET /api/v1/users/:id/sessions/messages?pageNo=1&pageSize=5
	}

	// 身份认证模块
	auth := v1.Group("/auth")
	{
		auth.POST("/sms", AuthHdl.SendSMSCode)                // POST /api/v1/auth/sms				发送短信验证码
		auth.POST("/email", AuthHdl.SendEmailCode)            // POST /api/v1/auth/email				发送邮箱验证码
		auth.POST("/login/password", AuthHdl.LoginByPassword) // POST /api/v1/auth/login/password 	手机号码/邮箱 + 密码登录
		auth.POST("/login/phone", AuthHdl.LoginByPhone)       // POST /api/v1/auth/login/phone 		手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册

		authedAuth := auth.Group("")
		authedAuth.Use(AuthRequiredMdl)
		authedAuth.POST("/logout", AuthHdl.Logout) // POST /api/v1/auth/logout			退出登录
		authedAuth.GET("/status", AuthHdl.Status)  // GET /api/v1/auth/status			检查登录状态

		authedAuth.POST("/password/update", AuthHdl.UpdatePassword) // POST /api/v1/auth/password/update	修改密码
		authedAuth.POST("/password/set", AuthHdl.SetPassword)       // POST /api/v1/auth/password/set	初始化密码
		authedAuth.GET("/password/status", AuthHdl.HasPassword)     // GET /api/v1/auth/password/status	查询密码状态

		authedAuth.GET("/auth_identity", AuthHdl.GetAuthIdentity) // GET /api/v1/auth/auth_identity	获取用户的身份认证
	}

	// 私信模块
	sessions := v1.Group("/sessions")
	sessions.Use(AuthRequiredMdl)
	{
		sessions.GET("", SessionHdl.List)               // GET /api/v1/sessions							获取当前登录用户会话列表
		sessions.POST("/:id/delete", SessionHdl.Delete) // DELETE /api/v1/sessions/:id						删除当前会话
	}

	// 帖子模块
	posts := v1.Group("/posts")
	{
		posts.GET("", PostHdl.List)                           // POST /api/v1/posts?pageNo=1&pageSize=10				按页获取帖子列表
		posts.GET("/top", PostHdl.Top)                        // GET /api/v1/posts/top								获取热门帖子榜单
		posts.GET("/tags", PostHdl.ListByTagAndPage)          // POST /api/v1/posts/tags?pageNo=1&pageSize=10&tag=go 根据标签按页获取帖子列表
		posts.GET("/:id", PostHdl.Detail)                     // GET /api/v1/posts/:id								获取帖子详情
		posts.GET("/:id/comments", PostHdl.ListCommentByPage) // GET /api/v1/posts/:id/comments?pageNo=1&pageSize=10	按页获取帖子评论
		posts.GET("/:id/comments/:cid", PostHdl.ListReplies)  // GET /api/v1/posts/:pid/comments/:cid?pageNo=1&pageSize=10	按页获取主评论回复

		//todo
		authedPosts := posts.Group("")
		authedPosts.Use(AuthRequiredMdl)
		authedPosts.POST("", PostHdl.CreatePost)            // POST /api/v1/posts 			创建帖子
		authedPosts.POST("/:id", PostHdl.Update)            // POST /api/v1/posts/:id 		更新帖子
		authedPosts.POST("/:id/delete", PostHdl.DeletePost) // POST /api/v1/posts/:id/delete	删除帖子

		authedPosts.POST("/:id/comments", PostHdl.CreateComment)             // POST /api/v1/posts/:id/comments 				创建评论
		authedPosts.POST("/:id/comments/:cid/delete", PostHdl.DeleteComment) // POST /api/v1/posts/:id/comments/:cid/delete 	删除评论
		authedPosts.GET("/:id/like", PostHdl.IfLike)                         // GET /api/v1/posts/:id/like					查询是否点赞了帖子
		authedPosts.POST("/:id/like", PostHdl.Like)                          // POST /api/v1/posts/:id/like					点赞帖子
		authedPosts.POST("/:id/unlike", PostHdl.Unlike)                      // POST /api/v1/posts/:id/unlike 				取消点赞帖子
	}

	// 搜索模块
	search := v1.Group("/search")
	search.Use(AuthRequiredMdl)
	{
		search.POST("", SearchHdl.Search)
	}

	// 智能体模块
	agent := v1.Group("/agent")
	agent.Use(AuthRequiredMdl)
	{
		agent.POST("/chat", AgentHdl.Chat) // POST /api/v1/agent/chat
	}

	// 抽奖模块
	v1.GET("/gifts", LotteryHdl.GetAllGifts) // GET /api/v1/gifts 获取所有奖品信息
	lottery := v1.Group("/lottery")
	lottery.Use(AuthRequiredMdl)
	{
		lottery.GET("/lucky", LotteryHdl.Lottery)  // GET /api/v1/lottery/lucky 抽奖
		lottery.POST("/giveup", LotteryHdl.GiveUp) // POST /api/v1/lottery/giveup 放弃
		lottery.POST("/pay", LotteryHdl.Pay)       // POST /api/v1/lottery/pay 支付
		lottery.GET("/result", LotteryHdl.Result)  // GET /api/v1/lottery/result 查询结果
	}

	// 即时聊天模块
	im := v1.Group("/ws")
	im.Use(AuthRequiredMdl)
	{
		im.GET("", WebsocketHdl.Connect) // GET /api/v1/ws
	}

	if err := engine.Run(BFFServiceConf.App.BackendAddr); err != nil {
		panic(err)
	}
}
