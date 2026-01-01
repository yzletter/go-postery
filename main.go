package main

import (
	"os"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yzletter/go-postery/handler"
	"github.com/yzletter/go-postery/infra/crontab"
	"github.com/yzletter/go-postery/infra/graceful_stop"
	infraMySQL "github.com/yzletter/go-postery/infra/mysql"
	infraRabbitMQ "github.com/yzletter/go-postery/infra/rabbitmq"
	infraRedis "github.com/yzletter/go-postery/infra/redis"
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
	slog.InitSlog("./logs/go_postery.log") // 初始化 slog
	crontab.InitCrontab()                  // 初始化 定时任务

	// 初始化 GracefulStop
	graceful_stop.NewGracefulStopBuilder().
		NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraMySQL.Close).AddFunc(infraRedis.Close).AddFunc(infraRabbitMQ.Close).
		Build()

	GormDB := infraMySQL.Init("./conf", "db", viper.YAML, "./logs") // 注册 MySQL
	RedisClient := infraRedis.Init("./conf", "cache", viper.YAML)   // 初始化 Redis
	RabbitMQ := infraRabbitMQ.Init("./conf", "mq", viper.YAML)      // 初始化 RabbitMQ

	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)   // 初始化 雪花算法
	PasswordHasher := security.NewBcryptPasswordHasher(0) // 初始化 密码哈希器
	JwtManager := security.NewJwtManager("123456")
	SmsClient := sms.NewAliyunSmsClient(os.Getenv("ALIYUN_AKID"), os.Getenv("ALIYUN_AKS")) // 初始化 短信服务商

	// 初始化 gin
	engine := gin.Default()

	// DAO 层
	UserDAO := dao.NewUserDAO(GormDB)
	PostDAO := dao.NewPostDAO(GormDB)
	CommentDAO := dao.NewCommentDAO(GormDB)
	LikeDAO := dao.NewLikeDAO(GormDB)
	FollowDAO := dao.NewFollowDAO(GormDB)
	TagDAO := dao.NewTagDAO(GormDB)
	MessageDAO := dao.NewMessageDAO(GormDB)
	SessionDAO := dao.NewSessionDAO(GormDB)

	// Cache 层
	UserCache := cache.NewUserCache(RedisClient)
	PostCache := cache.NewPostCache(RedisClient)
	CommentCache := cache.NewCommentCache(RedisClient)
	LikeCache := cache.NewLikeCache(RedisClient)
	FollowCache := cache.NewFollowCache(RedisClient)
	TagCache := cache.NewTagCache(RedisClient)
	MessageCache := cache.NewMessageCache(RedisClient)
	SessionCache := cache.NewSessionCache(RedisClient)
	SmsCache := cache.NewSmsCache(RedisClient)

	// Repository 层
	UserRepo := repository.NewUserRepository(UserDAO, UserCache)             // 注册 userRepo
	PostRepo := repository.NewPostRepository(PostDAO, PostCache)             // 注册 PostRepository
	CommentRepo := repository.NewCommentRepository(CommentDAO, CommentCache) // 注册 CommentRepository
	LikeRepo := repository.NewLikeRepository(LikeDAO, LikeCache)             // 注册 LikeRepository
	FollowRepo := repository.NewFollowRepository(FollowDAO, FollowCache)     // 注册 FollowRepository
	TagRepo := repository.NewTagRepository(TagDAO, TagCache)                 // 注册 TagRepository
	MessageRepo := repository.NewMessageRepository(MessageDAO, MessageCache)
	SessionRepo := repository.NewSessionRepository(SessionDAO, SessionCache)
	SmsRepo := repository.NewSmsRepository(SmsCache)

	// Service 层
	MetricSvc := service.NewMetricService()                                                           // 注册 MetricService
	RateLimitSvc := service.NewRateLimitService(RedisClient, time.Minute, 1000)                       // 注册 RateLimitService
	AuthSvc := service.NewAuthService(UserRepo, JwtManager, PasswordHasher, IDGenerator, RedisClient) // 注册 AuthService
	UserSvc := service.NewUserService(UserRepo, IDGenerator, PasswordHasher)                          // 注册 userSvc
	PostSvc := service.NewPostService(PostRepo, UserRepo, LikeRepo, TagRepo, IDGenerator)             // 注册 postSvc
	FollowSvc := service.NewFollowService(FollowRepo, UserRepo, IDGenerator)                          // 注册 FollowService
	CommentSvc := service.NewCommentService(CommentRepo, UserRepo, PostRepo, IDGenerator)             // 注册 commentService
	TagSvc := service.NewTagService(TagRepo, IDGenerator)                                             // 注册 TagService
	SessionSvc := service.NewSessionService(SessionRepo, MessageRepo, UserRepo, RabbitMQ, IDGenerator)
	WebsocketSvc := service.NewWebsocketService(SessionRepo, MessageRepo, UserRepo, RabbitMQ, IDGenerator)
	SmsSvc := service.NewSmsService(SmsClient, SmsRepo)

	// Handler 层
	AuthHdl := handler.NewAuthHandler(AuthSvc, SessionSvc)                // 注册 AuthHandler
	UserHdl := handler.NewUserHandler(UserSvc)                            // 注册 UserHandler
	PostHdl := handler.NewPostHandler(PostSvc, UserSvc, TagSvc)           // 注册 PostHandler
	CommentHdl := handler.NewCommentHandler(CommentSvc, UserSvc, PostSvc) // 注册 CommentHandler
	FollowHdl := handler.NewFollowHandler(FollowSvc, UserSvc)             // 注册 FollowHandler
	SessionHdl := handler.NewSessionHandler(SessionSvc)                   // 注册 SessionHandler
	WebsocketHdl := handler.NewWebsocketHandler(WebsocketSvc)             // 注册 WebsocketHandler
	SmsHdl := handler.NewSmsHandler(SmsSvc)

	// 中间件层
	AuthRequiredMdl := middleware.AuthRequiredMiddleware(AuthSvc, RedisClient) // AuthRequiredMdl 强制登录
	MetricMdl := middleware.MetricMiddleware(MetricSvc)                        // MetricMdl 用于 Prometheus 监控中间件
	RateLimitMdl := middleware.RateLimitMiddleware(RateLimitSvc)               // RateLimitMdl 限流中间件
	CorsMdl := // CorsMdl 跨域中间件
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:5173"}, // 允许域名跨域
			AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: true, // 是否允许携带 cookie 之类的用户认证信息
			ExposeHeaders:    []string{"Content-Length", "Authorization"},
			MaxAge:           12 * time.Hour,
		})

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
		// todo AuthHandler
		auth.POST("/register", AuthHdl.Register) // POST /api/v1/auth/register 	注册
		auth.POST("/login", AuthHdl.Login)       // POST /api/v1/auth/login 		登录

		auth.POST("/sms", SmsHdl.Send)                        // POST /api/v1/auth/sms			发送短信验证码
		auth.POST("/login/phone", AuthHdl.LoginByPhoneNumber) // POST /api/v1/auth/login 		手机号登录

		authedAuth := auth.Group("")
		authedAuth.Use(AuthRequiredMdl)
		authedAuth.POST("/logout", AuthHdl.Logout) // POST /api/v1/auth/logout	登出
		authedAuth.GET("/status", AuthHdl.Status)  // GET /api/v1/auth/status	检查状态
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
		me.POST("/password", UserHdl.ModifyPass)      // POST /api/v1/users/me/password 							修改密码
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
		sessions.GET("", SessionHdl.List)          // GET /api/v1/sessions								获取当前登录用户会话列表
		sessions.DELETE("/:id", SessionHdl.Delete) // DELETE /api/v1/sessions/:id						删除当前会话
	}

	im := v1.Group("/ws")
	im.Use(AuthRequiredMdl)
	{
		im.GET("", WebsocketHdl.Connect) // GET /api/v1/ws
	}

	if err := engine.Run("localhost:8765"); err != nil {
		panic(err)
	}
}
