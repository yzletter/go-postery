package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yzletter/go-postery/microservice-backend/bff/conf"
	"github.com/yzletter/go-postery/microservice-backend/bff/config"
	"github.com/yzletter/go-postery/microservice-backend/bff/grpc/client"
	"github.com/yzletter/go-postery/microservice-backend/bff/grpc/hub"
	handler2 "github.com/yzletter/go-postery/microservice-backend/bff/handler"
	"github.com/yzletter/go-postery/microservice-backend/bff/infra/crontab"
	infraEtcd "github.com/yzletter/go-postery/microservice-backend/bff/infra/etcd"
	"github.com/yzletter/go-postery/microservice-backend/bff/infra/graceful_stop"
	infraJaeger "github.com/yzletter/go-postery/microservice-backend/bff/infra/jaeger"
	infraRabbitMQ "github.com/yzletter/go-postery/microservice-backend/bff/infra/rabbitmq"
	infraRedis "github.com/yzletter/go-postery/microservice-backend/bff/infra/redis"
	infraSlog "github.com/yzletter/go-postery/microservice-backend/bff/infra/slog"
	middleware2 "github.com/yzletter/go-postery/microservice-backend/bff/middleware"
	service2 "github.com/yzletter/go-postery/microservice-backend/bff/service"
)

var (
	ServiceName  string // 微服务名
	GoPostery    string // GoPostery 公共配置前缀
	EtcdEndPoint string // etcd 地址
)

func main() {
	// 启动参数, 默认线上环境
	env := flag.String("env", "production", "运行环境: local/production")
	flag.Parse()

	// 本地测试
	if *env == "local" {
		ServiceName = "test_bff_service"
		GoPostery = "test_go_postery"
		EtcdEndPoint = "localhost:12379"
	} else {
		ServiceName = "bff_service"
		GoPostery = "go_postery"
		EtcdEndPoint = "172.16.131.223:2379"
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Remote Config Center
	EtcdClient := infraEtcd.Init([]string{EtcdEndPoint})                               // Init Etcd
	Config := config.LoadGlobalConfig(ctx, EtcdClient, ServiceName+"_", GoPostery+"_") // Get Config From Remote Config Center
	fmt.Printf("%s Init Config Success %+v\n", ServiceName, Config)

	// gRPC Common Infrastructure
	infraSlog.InitSlog(Config.Log)                                            // Init Slog
	TracerShutdown := infraJaeger.InitJaeger(ctx, Config.Jaeger, ServiceName) // Init JaegerTracer
	RedisClient := infraRedis.Init(Config.Redis)                              // 初始化 Redis

	// Infrastructure 层
	RabbitMQ := infraRabbitMQ.Init(Config.RabbitMQ) // 初始化 RabbitMQ

	// ServiceHub
	ETCDServiceHub := hub.NewEtcdServiceHub(Config.ServiceHub, EtcdClient, hub.NewRoundRobinLoadBalancer())
	ServiceHubProxy := hub.GetServiceHubProxy(ETCDServiceHub)
	ConnCenter := client.NewConnectionCenter(ServiceHubProxy)

	// GRPC Service 层
	AuthConn, err := ConnCenter.NewConnection(ctx, client.AuthServiceName)
	if err != nil {
		slog.Error("Init Auth gRPC Connection Failed", "error", err)
	}
	AuthServiceClient, err := client.NewAuthClient(AuthConn)
	if err != nil {
		slog.Error("Init Auth gRPC Client Failed", "error", err)
	}

	CodeConn, err := ConnCenter.NewConnection(ctx, client.CodeServiceName)
	if err != nil {
		slog.Error("Init Code gRPC Connection Failed", "error", err)
	}
	CodeServiceClient, err := client.NewCodeClient(CodeConn)
	if err != nil {
		slog.Error("Init Code gRPC Client Failed", "error", err)
	}

	UserConn, err := ConnCenter.NewConnection(ctx, client.UserServiceName)
	if err != nil {
		slog.Error("Init User gRPC Connection Failed", "error", err)
	}
	UserServiceClient, err := client.NewUserClient(UserConn)
	if err != nil {
		slog.Error("Init User gRPC Client Failed", "error", err)
	}

	PostConn, err := ConnCenter.NewConnection(ctx, client.PostServiceName)
	if err != nil {
		slog.Error("Init Post gRPC Connection Failed", "error", err)
	}
	PostServiceClient, err := client.NewPostClient(PostConn)
	if err != nil {
		slog.Error("Init Post gRPC Client Failed", "error", err)
	}

	SearchConn, err := ConnCenter.NewConnection(ctx, client.SearchServiceName)
	if err != nil {
		slog.Error("Init Search gRPC Connection Failed", "error", err)
	}
	SearchServiceClient, err := client.NewSearchClient(SearchConn)
	if err != nil {
		slog.Error("Init Search gRPC Client Failed", "error", err)
	}

	AgentConn, err := ConnCenter.NewConnection(ctx, client.AgentServiceName)
	if err != nil {
		slog.Error("Init Agent gRPC Connection Failed", "error", err)
	}
	AgentServiceClient, err := client.NewAgentClient(AgentConn)
	if err != nil {
		slog.Error("Init Agent gRPC Client Failed", "error", err)
	}

	LotteryConn, err := ConnCenter.NewConnection(ctx, client.LotteryServiceName)
	if err != nil {
		slog.Error("Init Lottery gRPC Connection Failed", "error", err)
	}
	LotteryServiceClient, err := client.NewLotteryClient(LotteryConn)
	if err != nil {
		slog.Error("Init Lottery gRPC Client Failed", "error", err)
	}

	SessionConn, err := ConnCenter.NewConnection(ctx, client.SessionServiceName)
	if err != nil {
		slog.Error("Init Session gRPC Connection Failed", "error", err)
	}
	SessionServiceClient, err := client.NewSessionClient(SessionConn)
	if err != nil {
		slog.Error("Init Session gRPC Client Failed", "error", err)
	}

	// Service 层
	MetricSvc := service2.NewMetricService()                                                              // 注册 MetricService
	RateLimitSvc := service2.NewRateLimitService(RedisClient, conf.RateLimitInterval, conf.RateLimitRate) // 注册 RateLimitService
	WebsocketSvc := service2.NewWebsocketService(SessionServiceClient, RabbitMQ)                          // 注册 WebsocketService

	// Handler 层
	PostHdl := handler2.NewPostHandler(PostServiceClient, UserServiceClient)                          // 注册 PostHandler
	AuthHdl := handler2.NewAuthHandler(AuthServiceClient, CodeServiceClient, UserServiceClient)       // 注册 AuthHandler
	UserHdl := handler2.NewUserHandler(UserServiceClient)                                             // 注册 UserHandler
	SearchHdl := handler2.NewSearchHandler(SearchServiceClient, PostServiceClient, UserServiceClient) // 注册 SearchHandler
	AgentHdl := handler2.NewAgentHandler(AgentServiceClient)                                          // 注册 AgentHandler
	LotteryHdl := handler2.NewLotteryHandler(LotteryServiceClient, UserServiceClient)                 // 注册 LotteryHandler
	SessionHdl := handler2.NewSessionHandler(SessionServiceClient, UserServiceClient)                 // 注册 SessionHandler
	WebsocketHdl := handler2.NewWebsocketHandler(WebsocketSvc)                                        // 注册 WebsocketHandler

	// 中间件层
	AuthRequiredMdl := middleware2.AuthRequiredMiddleware(AuthServiceClient) // AuthRequiredMdl 强制登录中间件
	MetricMdl := middleware2.MetricMiddleware(MetricSvc)                     // MetricMdl 用于 Prometheus 监控中间件
	RateLimitMdl := middleware2.RateLimitMiddleware(RateLimitSvc)            // RateLimitMdl 限流中间件
	CorsMdl := cors.New(cors.Config{                                         // CorsMdl 跨域中间件
		AllowOrigins:     []string{"http://" + Config.App.FrontendAddr}, // 允许域名跨域
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

	//// Prometheus
	//go func() {
	//	mux := http.NewServeMux()
	//	// Metric
	//	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) })
	//	metricAddr := ip + ":" + Config.Metric.Port
	//	if err := http.ListenAndServe(metricAddr, mux); err != nil {
	//		slog.Error("Metric Server Failed", "error", err)
	//	}
	//}()

	// Graceful Stop
	graceful_stop.NewGracefulStopBuilder().NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(cancel).AddFunc(TracerShutdown).
		Build()

	// 初始化 gin
	engine := gin.Default()
	engine.ContextWithFallback = true

	// 注册全局中间件
	engine.Use(
		middleware2.TracingMiddleware(ServiceName), // OpenTelemetry tracing 中间件
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

	// 用户模块
	users := v1.Group("/users")
	{
		users.GET("/:id", UserHdl.Profile)                    // GET /api/v1/users/:id								获取个人资料
		users.GET("/:id/posts", PostHdl.ListByPageAndUid)     // GET /api/v1/users/:id/posts?pageNo=1&pageSize=10		按页获取用户所发帖子
		users.GET("/top", UserHdl.Top)                        // GET /api/v1/users/top 								获取推荐关注
		users.POST("/presign", UserHdl.GetAvatarURL)          // POST /api/v1/users/presign 							预签名
		users.POST("/callback", UserHdl.UploadAvatarCallback) // POST /api/v1/users/callback 						回调

		// 个人模块
		me := users.Group("/me")
		me.Use(AuthRequiredMdl)
		me.POST("", UserHdl.ModifyProfile)          // POST /api/v1/users/me									修改个人资料
		me.GET("/upload", UserHdl.UploadAvatarSign) // GET /api/v1/users/me/upload							上传头像
		me.GET("/followers", UserHdl.ListFollowers) // GET /api/v1/users/me/followers?pageNo=1&pageSize=10	按页获取用户粉丝
		me.GET("/followees", UserHdl.ListFollowees) // GET /api/v1/users/me/followees?pageNo=1&pageSize=10 	按页获取用户关注的人

		// 关注模块
		follow := users.Group("/:id")
		follow.Use(AuthRequiredMdl)
		{
			follow.POST("/follow", UserHdl.Follow)     // POST /api/v1/users/:id/follow 	关注
			follow.POST("/unfollow", UserHdl.UnFollow) // Post /api/v1/users/:id/unfollow 取关
			follow.GET("/follow", UserHdl.IfFollow)    // GET /api/v1/users/:id/follow 	是否关注
		}

		// 私信模块
		chat := users.Group("/:id/sessions")
		chat.Use(AuthRequiredMdl)
		{
			chat.GET("", SessionHdl.GetSession)                 // GET /api/v1/users/:id/sessions									获取会话
			chat.GET("/messages", SessionHdl.GetHistoryMessage) // GET /api/v1/users/:id/sessions/messages?pageNo=1&pageSize=5		按页获取历史记录
		}
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

	if err := engine.Run(Config.App.BackendAddr); err != nil {
		panic(err)
	}
}
