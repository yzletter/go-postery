package main

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yzletter/go-postery/handler"
	"github.com/yzletter/go-postery/infra/crontab"
	"github.com/yzletter/go-postery/infra/security"
	"github.com/yzletter/go-postery/infra/slog"
	"github.com/yzletter/go-postery/infra/smooth"
	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/infra/viper"
	"github.com/yzletter/go-postery/middleware"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/service/ratelimit"

	infraMySQL "github.com/yzletter/go-postery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/infra/redis"
)

func main() {
	// Infra 层
	slog.InitSlog("./logs/go_postery.log") // 初始化 slog
	crontab.InitCrontab()                  // 初始化 定时任务
	smooth.InitSmoothExit()                // 初始化 优雅退出

	// Infra 层
	GormDB := infraMySQL.Init("./conf", "db", viper.YAML, "./logs") // 注册 MySQL
	RedisClient := infraRedis.Init("./conf", "redis", viper.YAML)   // 初始化 Redis
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)             // 初始化 雪花算法
	PasswordHasher := security.NewBcryptPasswordHasher(0)           // 初始化 密码哈希器
	JwtManager := security.NewJwtManager("123456")
	// 初始化 gin
	engine := gin.Default()

	// DAO 层
	UserDAO := dao.NewUserDAO(GormDB)
	PostDAO := dao.NewPostDAO(GormDB)
	CommentDAO := dao.NewCommentDAO(GormDB)
	LikeDAO := dao.NewLikeDAO(GormDB)
	FollowDAO := dao.NewFollowDAO(GormDB)
	TagDAO := dao.NewTagDAO(GormDB)

	// Cache 层
	UserCache := cache.NewUserCache(RedisClient)
	PostCache := cache.NewPostCache(RedisClient)
	CommentCache := cache.NewCommentCache(RedisClient)
	LikeCache := cache.NewLikeCache(RedisClient)
	FollowCache := cache.NewFollowCache(RedisClient)
	TagCache := cache.NewTagCache(RedisClient)

	// Repository 层
	UserRepo := repository.NewUserRepository(UserDAO, UserCache)             // 注册 UserRepo
	PostRepo := repository.NewPostRepository(PostDAO, PostCache)             // 注册 PostRepository
	CommentRepo := repository.NewCommentRepository(CommentDAO, CommentCache) // 注册 CommentRepository
	LikeRepo := repository.NewLikeRepository(LikeDAO, LikeCache)             // 注册 LikeRepository
	FollowRepo := repository.NewFollowRepository(FollowDAO, FollowCache)     // 注册 FollowRepository
	TagRepo := repository.NewTagRepository(TagDAO, TagCache)                 // 注册 TagRepository

	// Service 层
	UserSvc := service.NewUserService(UserRepo, IDGenerator, PasswordHasher) // 注册 UserService
	AuthSvc := service.NewAuthService(UserRepo, JwtManager, PasswordHasher, IDGenerator, RedisClient)

	// Handler 层
	AuthHdl := handler.NewAuthHandler(AuthSvc)
	UserHdl := handler.NewUserHandler(UserSvc) // 注册 UserHandler

	// 中间件层
	AuthRequiredMdl := middleware.AuthRequiredMiddleware(AuthSvc, RedisClient) // AuthRequiredMdl 强制登录

	// todo 待重构
	// todo 待重构
	// todo 待重构
	// todo 待重构
	// todo 待重构
	PostSvc := service.NewPostService(PostRepo, UserRepo, LikeRepo, TagRepo) // 注册 PostService
	FollowSvc := service.NewFollowService(FollowRepo, UserRepo)              // 注册 FollowService
	CommentSvc := service.NewCommentService(CommentRepo, UserRepo, PostRepo) // 注册 CommentService
	TagSvc := service.NewTagService(TagRepo)                                 // 注册 TagService

	MetricSvc := service.NewMetricService()                                       // 注册 MetricService
	RateLimitSvc := ratelimit.NewRateLimitService(RedisClient, time.Minute, 1000) // 注册 RateLimitService

	// Handler 层
	PostHdl := handler.NewPostHandler(PostSvc, UserSvc, TagSvc)           // 注册 PostHandler
	CommentHdl := handler.NewCommentHandler(CommentSvc, UserSvc, PostSvc) // 注册 CommentHandler
	FollowHdl := handler.NewFollowHandler(FollowSvc, UserSvc)

	// 中间件层
	MetricMdl := middleware.MetricMiddleware(MetricSvc)          // MetricMdl 用于 Prometheus 监控中间件
	RateLimitMdl := middleware.RateLimitMiddleware(RateLimitSvc) // RateLimitMdl 限流中间件

	CorsMdl := cors.New(cors.Config{ // CorsMdl 跨域中间件
		AllowOrigins:  []string{"http://localhost:5173"}, // 允许域名跨域
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length", "x-jwt-token"},

		// 判断来源的函数
		AllowOriginFunc: func(origin string) bool {
			if strings.Contains(origin, "http://localhost") { // 开发环境
				return true
			}
			return strings.Contains(origin, "gopostery.com")
		},
		MaxAge: 12 * time.Hour,
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

	// /api/v1
	// 身份认证模块
	auth := v1.Group("/auth")
	{
		// todo AuthHandler
		auth.POST("/register", AuthHdl.Register) // POST /api/v1/auth/register
		auth.POST("/login", AuthHdl.Login)       // POST /api/v1/auth/login

		authedAuth := auth.Group("")
		authedAuth.Use(AuthRequiredMdl)
		authedAuth.POST("/logout", UserHdl.Logout) // POST /api/v1/auth/logout
	}

	// 用户模块
	users := v1.Group("/users")
	{
		users.GET("/:id", UserHdl.Profile)
		users.GET("/:id/posts", PostHdl.ListByUid) // 替代 /posts/user/:uid

		// 个人模块
		me := users.Group("/me")
		me.Use(AuthRequiredMdl)
		me.PATCH("", UserHdl.ModifyProfile)
		me.PATCH("/password", UserHdl.ModifyPass)
		me.GET("/followers", FollowHdl.ListFollowers)
		me.GET("/followees", FollowHdl.ListFollowees)

		// 关注模块
		follow := users.Group("/:id/follow")
		follow.Use(AuthRequiredMdl)
		{
			follow.PUT("", FollowHdl.Follow)      // 关注
			follow.DELETE("", FollowHdl.UnFollow) // 取关
			follow.GET("", FollowHdl.IfFollow)    // 是否关注
		}
	}

	// 帖子模块
	posts := v1.Group("/posts")
	{
		posts.GET("", PostHdl.List) // 支持 ?pageNo&pageSize&tagId&uid...
		posts.GET("/:id", PostHdl.Detail)
		posts.GET("/:id/comments", CommentHdl.List)

		authedPosts := posts.Group("")
		authedPosts.Use(AuthRequiredMdl)
		authedPosts.POST("", PostHdl.Create) // 替代 /me/posts/new
		authedPosts.PATCH("/:id", PostHdl.Update)
		authedPosts.DELETE("/:id", PostHdl.Delete)

		authedPosts.POST("/:id/comments", CommentHdl.Create)         // 替代 /me/comments/new
		authedPosts.DELETE("/:pid/comments/:cid", CommentHdl.Delete) // 或 DELETE /comments/:cid
		authedPosts.PUT("/:id/likes", PostHdl.Like)                  // 幂等点赞
		authedPosts.DELETE("/:id/likes", PostHdl.Unlike)             // 幂等取消
	}

	// 管理员模块
	admin := v1.Group("/admin")
	{
		admin.Use(AuthRequiredMdl, AdminRequiredMdl)
	}

	if err := engine.Run("localhost:8765"); err != nil {
		panic(err)
	}
}
