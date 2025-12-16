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
	"github.com/yzletter/go-postery/middleware"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
	"github.com/yzletter/go-postery/router"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/service/ratelimit"

	infraMySQL "github.com/yzletter/go-postery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/infra/redis"
)

func main() {
	// Infra 层
	infraMySQL.Init("./conf", "db", viper.YAML, "./logs") // 注册 MySQL
	infraRedis.Init("./conf", "redis", viper.YAML)        // 注册 Redis
	slog.InitSlog("./logs/go_postery.log")                // 初始化 slog
	crontab.InitCrontab()                                 // 初始化 定时任务
	smooth.InitSmoothExit()                               // 初始化 优雅退出
	snowflake.Init(0)                                     // 初始化 雪花算法

	// 初始化 gin
	engine := gin.Default()

	// DAO 层
	UserDAO := dao.NewUserDAO(infraMySQL.GetDB())
	PostDAO := dao.NewPostDAO(infraMySQL.GetDB())
	CommentDAO := dao.NewCommentDAO(infraMySQL.GetDB())
	LikeDAO := dao.NewLikeDAO(infraMySQL.GetDB())
	FollowDAO := dao.NewFollowDAO(infraMySQL.GetDB())
	TagDAO := dao.NewTagDAO(infraMySQL.GetDB())

	// Cache 层
	UserCache := cache.NewUserCache(infraRedis.GetRedis())
	PostCache := cache.NewPostCache(infraRedis.GetRedis())
	CommentCache := cache.NewCommentCache(infraRedis.GetRedis())
	LikeCache := cache.NewLikeCache(infraRedis.GetRedis())
	FollowCache := cache.NewFollowCache(infraRedis.GetRedis())
	TagCache := cache.NewTagCache(infraRedis.GetRedis())

	// Repository 层
	UserRepo := repository.NewUserRepository(UserDAO, UserCache)             // 注册 UserRepository
	PostRepo := repository.NewPostRepository(PostDAO, PostCache)             // 注册 PostRepository
	CommentRepo := repository.NewCommentRepository(CommentDAO, CommentCache) // 注册 CommentRepository
	LikeRepo := repository.NewLikeRepository(LikeDAO, LikeCache)             // 注册 LikeRepository
	FollowRepo := repository.NewFollowRepository(FollowDAO, FollowCache)     // 注册 FollowRepository
	TagRepo := repository.NewTagRepository(TagDAO, TagCache)                 // 注册 TagRepository

	// Service 层
	UserSvc := service.NewUserService(UserRepo)                              // 注册 userService
	PostSvc := service.NewPostService(PostRepo, UserRepo, LikeRepo, TagRepo) // 注册 postService
	FollowSvc := service.NewFollowService(FollowRepo, UserRepo)              // 注册 followService
	CommentSvc := service.NewCommentService(CommentRepo, UserRepo, PostRepo) // 注册 CommentService
	TagSvc := service.NewTagService(TagRepo)                                 // 注册 tagService

	JwtSvc := service.NewJwtService(infraRedis.GetRedis(), "123456")                        // 注册 JwtService
	MetricSvc := service.NewMetricService()                                                 // 注册 MetricService
	RateLimitSvc := ratelimit.NewRateLimitService(infraRedis.GetRedis(), time.Minute, 1000) // 注册 RateLimitService

	// Handler 层
	UserHdl := handler.NewUserHandler(UserSvc, JwtSvc)                    // 注册 UserHandler
	PostHdl := handler.NewPostHandler(PostSvc, UserSvc, TagSvc)           // 注册 PostHandler
	CommentHdl := handler.NewCommentHandler(CommentSvc, UserSvc, PostSvc) // 注册 CommentHandler
	FollowHdl := handler.NewFollowHandler(FollowSvc, UserSvc)

	// 中间件层
	AuthRequiredMdl := middleware.AuthRequiredMiddleware(JwtSvc) // AuthRequiredMdl 强制登录
	MetricMdl := middleware.MetricMiddleware(MetricSvc)          // MetricMdl 用于 Prometheus 监控中间件
	RateLimitMdl := middleware.RateLimitMiddleware(RateLimitSvc) // RateLimitMdl 限流中间件
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

	// 运维接口
	engine.GET("/metrics", func(ctx *gin.Context) { // Prometheus 访问的接口
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request) // 固定写法
	})

	// 业务接口
	api := engine.Group("/api")
	v1 := api.Group("/v1")

	// localhost:8765/api/v1
	// 身份认证模块
	auth := v1.Group("/auth")
	{
		// todo AuthHandler
		auth.POST("/register", UserHdl.Register) // localhost:8765/api/v1/auth/register
		auth.POST("/login", UserHdl.Login)

		authedAuth := auth.Group("")
		authedAuth.Use(AuthRequiredMdl)
		authedAuth.POST("/logout", UserHdl.Logout)
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
