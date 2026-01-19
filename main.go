package main

import (
	"context"
	"os"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/handler"
	"github.com/yzletter/go-postery/infra/crontab"
	"github.com/yzletter/go-postery/infra/email"
	"github.com/yzletter/go-postery/infra/graceful_stop"
	infraKafka "github.com/yzletter/go-postery/infra/kafka"
	infraLLM "github.com/yzletter/go-postery/infra/llm"
	infraMySQL "github.com/yzletter/go-postery/infra/mysql"
	infraQdarant "github.com/yzletter/go-postery/infra/qdrant"
	infraRabbitMQ "github.com/yzletter/go-postery/infra/rabbitmq"
	infraRedis "github.com/yzletter/go-postery/infra/redis"
	infraRocketMQ "github.com/yzletter/go-postery/infra/rocketmq"
	"github.com/yzletter/go-postery/infra/security"
	"github.com/yzletter/go-postery/infra/slog"
	"github.com/yzletter/go-postery/infra/sms"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/infra/viper"
	"github.com/yzletter/go-postery/middleware"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
	"github.com/yzletter/go-postery/service"
)

func main() {
	// Infra 层
	slog.InitSlog(conf.LogFilePath) // 初始化 slog

	MySQLGormDB := infraMySQL.Init("./conf", "db", viper.YAML, "./logs")                                                 // 初始化 MySQL
	QdrantClient := infraQdarant.Init("./conf", "db", viper.YAML)                                                        // 初始化 Qdrant
	ArkEmbedder := infraLLM.NewArkEmbedder(context.Background(), "doubao-embedding-vision-250615", os.Getenv("ARK_KEY")) // 初始化火山引擎向量模型
	ArkChatModel := infraLLM.NewArkModel(context.Background(), "doubao-seed-1-8-251228", os.Getenv("ARK_KEY"))
	RedisClient := infraRedis.Init("./conf", "cache", viper.YAML)                                       // 初始化 Redis
	RabbitMQ := infraRabbitMQ.Init("./conf", "mq", viper.YAML)                                          // 初始化 RabbitMQ
	RocketMQ := infraRocketMQ.Init(conf.RocketProxyEndpoint)                                            // 初始化 RocketMQ
	KafkaProducer := infraKafka.InitProducer([]string{conf.KafkaEndpoint})                              // 初始化 Kafka 生产方
	SessionKafkaConsumer := infraKafka.InitConsumer([]string{conf.KafkaEndpoint}, "session", "session") // 初始化 Session 模块 Kafka 消费方
	FollowKafkaConsumer := infraKafka.InitConsumer([]string{conf.KafkaEndpoint}, "follow", "follow")    // 初始化 Follow 模块 Kafka 消费方
	QdrantKafkaConsumer := infraKafka.InitConsumer([]string{conf.KafkaEndpoint}, "upsert_qdrant", "agent_qdrant")
	AgentKafkaConsumer := infraKafka.InitConsumer([]string{conf.KafkaEndpoint}, "index_document", "agent_document")

	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)   // 初始化 雪花算法
	PasswordHasher := security.NewBcryptPasswordHasher(0) // 初始化 密码哈希器
	JwtManager := security.NewJwtManager(conf.JwtTokenKey)
	EmailManager := email.NewEmailManager(conf.EmailFrom, os.Getenv(conf.EmailAuthCode), conf.EmailSubject, conf.EmailExpireMin, conf.AppName, conf.EmailYear, conf.Address)
	SmsClient := sms.NewAliyunSmsClient(os.Getenv(conf.AliyunAccessTokenKeyID), os.Getenv(conf.AliyunAccessTokenKeySecret)) // 初始化 短信服务商

	// DAO 层
	UserDAO := dao.NewUserDAO(MySQLGormDB)
	PostDAO := dao.NewPostDAO(MySQLGormDB)
	CommentDAO := dao.NewCommentDAO(MySQLGormDB)
	LikeDAO := dao.NewLikeDAO(MySQLGormDB)
	FollowDAO := dao.NewFollowDAO(MySQLGormDB)
	TagDAO := dao.NewTagDAO(MySQLGormDB)
	MessageDAO := dao.NewMessageDAO(MySQLGormDB)
	SessionDAO := dao.NewSessionDAO(MySQLGormDB)
	OrderDAO := dao.NewOrderDAO(MySQLGormDB)
	GiftDAO := dao.NewGiftDAO(MySQLGormDB)
	AuthDAO := dao.NewAuthDAO(MySQLGormDB)
	AgentDAO := dao.NewAgentDAO(MySQLGormDB, QdrantClient, ArkEmbedder.GetInternal())

	// Cache 层
	UserCache := cache.NewUserCache(RedisClient)
	PostCache := cache.NewPostCache(RedisClient)
	CommentCache := cache.NewCommentCache(RedisClient)
	LikeCache := cache.NewLikeCache(RedisClient)
	FollowCache := cache.NewFollowCache(RedisClient)
	TagCache := cache.NewTagCache(RedisClient)
	MessageCache := cache.NewMessageCache(RedisClient)
	SessionCache := cache.NewSessionCache(RedisClient)
	CodeCache := cache.NewCodeCache(RedisClient)
	OrderCache := cache.NewOrderCache(RedisClient)
	GiftCache := cache.NewGiftCache(RedisClient)
	AuthCache := cache.NewAuthCache(RedisClient)

	// Repository 层
	UserRepo := repository.NewUserRepository(UserDAO, UserCache)             // 注册 userRepo
	PostRepo := repository.NewPostRepository(PostDAO, PostCache)             // 注册 PostRepository
	CommentRepo := repository.NewCommentRepository(CommentDAO, CommentCache) // 注册 CommentRepository
	LikeRepo := repository.NewLikeRepository(LikeDAO, LikeCache)             // 注册 LikeRepository
	FollowRepo := repository.NewFollowRepository(FollowDAO, FollowCache)     // 注册 FollowRepository
	TagRepo := repository.NewTagRepository(TagDAO, TagCache)                 // 注册 TagRepository
	MessageRepo := repository.NewMessageRepository(MessageDAO, MessageCache) // 注册 MessageRepository
	SessionRepo := repository.NewSessionRepository(SessionDAO, SessionCache) // 注册 SessionRepository
	OrderRepo := repository.NewOrderRepository(OrderDAO, OrderCache)         // 注册 OrderRepository
	GiftRepo := repository.NewGiftRepository(GiftDAO, GiftCache)             // 注册 GiftRepository
	AuthRepo := repository.NewAuthRepository(AuthDAO, AuthCache)
	CodeRepo := repository.NewCodeRepository(CodeCache)
	AgentRepo := repository.NewAgentRepository(AgentDAO)

	// Service 层
	CodeSvc := service.NewCodeService(CodeRepo, EmailManager, SmsClient)
	MetricSvc := service.NewMetricService()                                                                                  // 注册 MetricService
	RateLimitSvc := service.NewRateLimitService(RedisClient, conf.RateLimitInterval, conf.RateLimitRate)                     // 注册 RateLimitService
	AuthSvc := service.NewAuthService(CodeSvc, AuthRepo, UserRepo, JwtManager, PasswordHasher, IDGenerator)                  // 注册 AuthService
	UserSvc := service.NewUserService(UserRepo, IDGenerator, PasswordHasher)                                                 // 注册 userSvc
	PostSvc := service.NewPostService(PostRepo, UserRepo, LikeRepo, TagRepo, IDGenerator)                                    // 注册 postSvc
	FollowSvc := service.NewFollowService(FollowRepo, UserRepo, FollowKafkaConsumer, IDGenerator)                            // 注册 FollowService
	CommentSvc := service.NewCommentService(CommentRepo, UserRepo, PostRepo, IDGenerator)                                    // 注册 commentService
	TagSvc := service.NewTagService(TagRepo, IDGenerator)                                                                    // 注册 TagService
	SessionSvc := service.NewSessionService(SessionRepo, MessageRepo, UserRepo, RabbitMQ, SessionKafkaConsumer, IDGenerator) // 注册 SessionService
	WebsocketSvc := service.NewWebsocketService(SessionRepo, MessageRepo, UserRepo, RabbitMQ, IDGenerator)                   // 注册 WebsocketService
	LotterySvc := service.NewLotteryService(OrderRepo, GiftRepo, UserRepo, RocketMQ, IDGenerator)                            // 注册 LotteryService
	AgentSvc := service.NewAgentService(AgentRepo, PostRepo, CommentRepo, AgentKafkaConsumer, QdrantKafkaConsumer, ArkEmbedder, ArkChatModel, IDGenerator)

	// 开启协程
	ctx, cancel := context.WithCancel(context.Background())

	go infraMySQL.ScanOutbox(ctx, KafkaProducer)    // 开启扫表发消息协程
	go AgentSvc.StartChunkDocConsumer(ctx)          // 开启切分文档协程
	go AgentSvc.StartUpsertQdrantConsumer(ctx)      // 开启切分文档协程
	go FollowSvc.StartInitUserScoreConsumer(ctx)    // 开启修改用户分数的消息协程
	go SessionSvc.StartSessionRegisterConsumer(ctx) // 开启协程注册新用户聊天功能
	go LotterySvc.StartLotteryOrderConsumer(ctx)    // 开启协程
	LotterySvc.InitCacheInventory(ctx)              // 初始化缓存库存

	// Handler 层
	AuthHdl := handler.NewAuthHandler(AuthSvc, CodeSvc)                   // 注册 AuthHandler
	UserHdl := handler.NewUserHandler(UserSvc)                            // 注册 UserHandler
	PostHdl := handler.NewPostHandler(PostSvc, UserSvc, TagSvc)           // 注册 PostHandler
	CommentHdl := handler.NewCommentHandler(CommentSvc, UserSvc, PostSvc) // 注册 CommentHandler
	FollowHdl := handler.NewFollowHandler(FollowSvc, UserSvc)             // 注册 FollowHandler
	SessionHdl := handler.NewSessionHandler(SessionSvc)                   // 注册 SessionHandler
	WebsocketHdl := handler.NewWebsocketHandler(WebsocketSvc)             // 注册 WebsocketHandler
	LotteryHdl := handler.NewLotteryHandler(LotterySvc)                   // 注册 LotteryHandler
	AgentHdl := handler.NewAgentHandler(AgentSvc)                         // 注册 AgentHandler

	// 中间件层
	AuthRequiredMdl := middleware.AuthRequiredMiddleware(AuthSvc) // AuthRequiredMdl 强制登录中间件
	MetricMdl := middleware.MetricMiddleware(MetricSvc)           // MetricMdl 用于 Prometheus 监控中间件
	RateLimitMdl := middleware.RateLimitMiddleware(RateLimitSvc)  // RateLimitMdl 限流中间件
	CorsMdl := cors.New(cors.Config{                              // CorsMdl 跨域中间件
		AllowOrigins:     []string{conf.FrontendEndPoint}, // 允许域名跨域
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true, // 是否允许携带 cookie 之类的用户认证信息
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		MaxAge:           12 * time.Hour,
	})

	// 初始化 Crontab
	crontab.NewCrontabBuilder().
		AddFuncWithSpec("*/10 * * * *", infraMySQL.Ping).
		AddFuncWithSpec("*/10 * * * *", infraRedis.Ping).
		Build()

	// 初始化 GracefulStop
	graceful_stop.NewGracefulStopBuilder().
		NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(cancel).                                                                     // 关协程
		AddFunc(infraRabbitMQ.Close).AddFunc(infraRocketMQ.Close).AddFunc(infraKafka.Close). // 关消息队列
		AddFunc(infraMySQL.Close).AddFunc(infraRedis.Close).AddFunc(infraQdarant.Close).     // 关数据库
		Build()

	// 初始化 gin
	engine := gin.Default()

	// 注册全局中间件
	engine.Use(
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
		users.GET("/:id", UserHdl.Profile)                // GET /api/v1/users/:id									获取个人资料
		users.GET("/:id/posts", PostHdl.ListByPageAndUid) // GET /api/v1/users/:id/posts?pageNo=1&pageSize=10		按页获取用户所发帖子
		users.GET("/top", UserHdl.Top)                    // GET /api/v1/users/top 									获取推荐关注

		// 个人模块
		me := users.Group("/me")
		me.Use(AuthRequiredMdl)
		me.POST("", UserHdl.ModifyProfile)            // POST /api/v1/users/me									修改个人资料
		me.GET("/followers", FollowHdl.ListFollowers) // GET /api/v1/users/me/followers?pageNo=1&pageSize=10		按页获取用户粉丝
		me.GET("/followees", FollowHdl.ListFollowees) // GET /api/v1/users/me/followees?pageNo=1&pageSize=10 	按页获取用户关注的人

		// 关注模块
		follow := users.Group("/:id/follow")
		follow.Use(AuthRequiredMdl)
		{
			follow.POST("", FollowHdl.Follow)     // POST /api/v1/users/:id/follow 		关注
			follow.DELETE("", FollowHdl.UnFollow) // DELETE /api/v1/users/:id/follow 	取关
			follow.GET("", FollowHdl.IfFollow)    // GET /api/v1/users/:id/follow 		是否关注
		}

		// 私信模块
		chat := users.Group("/:id/sessions")
		chat.Use(AuthRequiredMdl)
		{
			chat.GET("", SessionHdl.GetSession)                 // GET /api/v1/users/:id/sessions									获取会话
			chat.GET("/messages", SessionHdl.GetHistoryMessage) // GET /api/v1/users/:id/sessions/messages?pageNo=1&pageSize=5		按页获取历史记录
		}
	}

	// 帖子模块
	posts := v1.Group("/posts")
	{
		posts.GET("", PostHdl.List)                             // POST /api/v1/posts?pageNo=1&pageSize=10				按页获取帖子列表
		posts.GET("/top", PostHdl.Top)                          // GET /api/v1/posts/top								获取热门帖子榜单
		posts.GET("/tags", PostHdl.ListByTagAndPage)            // POST /api/v1/posts/tags?pageNo=1&pageSize=10&tag=go 根据标签按页获取帖子列表
		posts.GET("/:id", PostHdl.Detail)                       // GET /api/v1/posts/:id								获取帖子详情
		posts.GET("/:id/comments", CommentHdl.ListByPage)       // GET /api/v1/posts/:id/comments?pageNo=1&pageSize=10	按页获取帖子评论
		posts.GET("/:id/comments/:cid", CommentHdl.ListReplies) // GET /api/v1/posts/:pid/comments/:cid?pageNo=1&pageSize=10	按页获取主评论回复

		//todo
		authedPosts := posts.Group("")
		authedPosts.Use(AuthRequiredMdl)
		authedPosts.POST("", PostHdl.Create)       // POST /api/v1/posts 		创建帖子
		authedPosts.POST("/:id", PostHdl.Update)   // POST /api/v1/posts/:id 	更新帖子
		authedPosts.DELETE("/:id", PostHdl.Delete) // DELETE /api/v1/posts/:id 	删除帖子

		authedPosts.POST("/:id/comments", CommentHdl.Create)        // POST /api/v1/posts/:id/comments 创建评论
		authedPosts.DELETE("/:id/comments/:cid", CommentHdl.Delete) // DELETE /api/v1/posts/:id/comments/:cid 删除评论
		authedPosts.GET("/:id/likes", PostHdl.IfLike)               // GET /api/v1/posts/:id/likes	查询是否点赞了帖子
		authedPosts.POST("/:id/likes", PostHdl.Like)                // POST /api/v1/posts/:id/likes	点赞帖子
		authedPosts.DELETE("/:id/likes", PostHdl.Unlike)            // DELETE /api/v1/posts/:id/likes 取消点赞帖子
	}

	// 私信模块
	sessions := v1.Group("/sessions")
	sessions.Use(AuthRequiredMdl)
	{
		sessions.GET("", SessionHdl.List)          // GET /api/v1/sessions							获取当前登录用户会话列表
		sessions.DELETE("/:id", SessionHdl.Delete) // DELETE /api/v1/sessions/:id						删除当前会话
	}

	// 即时聊天模块
	im := v1.Group("/ws")
	im.Use(AuthRequiredMdl)
	{
		im.GET("", WebsocketHdl.Connect) // GET /api/v1/ws
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

	agent := v1.Group("/agent")
	agent.Use(AuthRequiredMdl)
	{
		agent.POST("/chat", AgentHdl.Chat) // POST /api/v1/agent/chat
	}

	if err := engine.Run("localhost:8765"); err != nil {
		panic(err)
	}
}
