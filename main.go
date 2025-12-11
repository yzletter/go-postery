package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yzletter/go-postery/handler"
	"github.com/yzletter/go-postery/infra/crontab"
	"github.com/yzletter/go-postery/infra/slog"
	"github.com/yzletter/go-postery/infra/smooth"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/infra/viper"
	"github.com/yzletter/go-postery/service/ratelimit"

	infraMySQL "github.com/yzletter/go-postery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/infra/redis"

	"github.com/yzletter/go-postery/middleware"

	commentRepository "github.com/yzletter/go-postery/repository/comment"
	postRepository "github.com/yzletter/go-postery/repository/post"
	userRepository "github.com/yzletter/go-postery/repository/user"

	"github.com/yzletter/go-postery/service"
)

func main() {
	// Infra 层
	infraMySQL.Init("./conf", "db", viper.YAML, "./log") // 注册 MySQL
	infraRedis.Init("./conf", "redis", viper.YAML)       // 注册 Redis
	slog.InitSlog("./log/go_postery.log")                // 初始化 slog
	crontab.InitCrontab()                                // 初始化 定时任务
	smooth.InitSmoothExit()                              // 初始化 优雅退出
	snowflake.Init(0)                                    // 初始化 雪花算法
	// 初始化 gin
	engine := gin.Default()

	// Repository 层
	UserDBRepo := userRepository.NewUserDBRepository(infraMySQL.GetDB())                   // 注册 UserDBRepo
	PostDBRepo := postRepository.NewPostDBRepository(infraMySQL.GetDB())                   // 注册 PostDBRepo
	CommentDBRepo := commentRepository.NewCommentDBRepository(infraMySQL.GetDB())          // 注册 CommentDBRepo
	UserCacheRepo := userRepository.NewUserCacheRepository(infraRedis.GetRedis())          // 注册 UserCacheRepo
	PostCacheRepo := postRepository.NewPostCacheRepository(infraRedis.GetRedis())          // 注册 PostCacheRepo
	CommentCacheRepo := commentRepository.NewCommentCacheRepository(infraRedis.GetRedis()) // 注册 CommentCacheRepo

	// Service 层
	JwtSvc := service.NewJwtService("123456")                                                        // 注册 JwtSvc
	MetricSvc := service.NewMetricService()                                                          // 注册 MetricSvc
	AuthSvc := service.NewAuthService(infraRedis.GetRedis(), JwtSvc)                                 // 注册 AuthSvc
	RateLimitSvc := ratelimit.NewRateLimitService(infraRedis.GetRedis(), time.Minute, 500)           // 注册 RateLimitSvc
	UserSvc := service.NewUserService(UserDBRepo, UserCacheRepo)                                     // 注册 UserSvc
	PostSvc := service.NewPostService(PostDBRepo, PostCacheRepo, UserDBRepo)                         // 注册 PostSvc
	CommentSvc := service.NewCommentService(CommentDBRepo, CommentCacheRepo, UserDBRepo, PostDBRepo) // 注册 CommentSvc

	// Handler 层
	UserHdl := handler.NewUserHandler(AuthSvc, JwtSvc, UserSvc)           // 注册 UserHandler
	PostHdl := handler.NewPostHandler(PostSvc, UserSvc)                   // 注册 PostHandler
	CommentHdl := handler.NewCommentHandler(CommentSvc, UserSvc, PostSvc) // 注册 CommentHandler

	// 中间件层
	AuthRequiredMdl := middleware.AuthRequiredMiddleware(AuthSvc) // AuthRequiredMdl 强制登录
	AuthOptionalMdl := middleware.AuthOptionalMiddleware(AuthSvc) // AuthOptionalMdl 非强制要求登录
	MetricMdl := middleware.MetricMiddleware(MetricSvc)           // MetricMdl 用于 Prometheus 监控中间件
	RateLimitMdl := middleware.RateLimitMiddleware(RateLimitSvc)  // RateLimitMdl 限流中间件
	CorsMdl := cors.New(cors.Config{ // CorsMdl 跨域中间件
		AllowOrigins:     []string{"http://localhost:5173"}, // 允许域名跨域
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})

	// 注册全局中间件
	engine.Use(
		CorsMdl,      // CorsMdl 跨域中间件
		MetricMdl,    // Prometheus 监控中间件
		RateLimitMdl, // 限流中间件
	)

	// 定义路由
	engine.GET("/metrics", func(ctx *gin.Context) { // Prometheus 访问的接口
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request) // 固定写法
	})

	// 用户模块
	engine.POST("/register/submit", UserHdl.Register) // 用户注册
	engine.POST("/login/submit", UserHdl.Login)       // 用户登录
	engine.GET("/logout", UserHdl.Logout)             // 用户退出
	engine.GET("/profile/:id", UserHdl.Profile)       // 用户资料
	// 强制登录
	engine.POST("/modify_pass/submit", AuthRequiredMdl, UserHdl.ModifyPass)       // 修改密码
	engine.POST("/modify_profile/submit", AuthRequiredMdl, UserHdl.ModifyProfile) // 修改个人资料

	// 帖子模块
	engine.GET("/posts", PostHdl.List)               // 获取帖子列表
	engine.GET("/posts/:pid", PostHdl.Detail)        // 获取帖子详情
	engine.GET("/posts_uid/:uid", PostHdl.ListByUid) // 获取目标用户发布的帖子
	// 强制登录
	engine.POST("/posts/new", AuthRequiredMdl, PostHdl.Create)       // 创建帖子
	engine.GET("/posts/delete/:id", AuthRequiredMdl, PostHdl.Delete) // 删除帖子
	engine.POST("/posts/update", AuthRequiredMdl, PostHdl.Update)    // 修改帖子
	// 非强制要求登录
	engine.GET("/posts/belong", AuthOptionalMdl, PostHdl.Belong) // 查询帖子是否归属当前登录用户

	// 评论模块
	engine.GET("/comment/list/:post_id", CommentHdl.List) // 列出评论
	// 强制登录
	engine.POST("/comment/new", AuthRequiredMdl, CommentHdl.Create)       // 创建评论
	engine.GET("/comment/delete/:id", AuthRequiredMdl, CommentHdl.Delete) // 删除评论
	engine.GET("/comment/belong", AuthRequiredMdl, CommentHdl.Belong)     // 删除评论

	if err := engine.Run("localhost:8765"); err != nil {
		panic(err)
	}
}
