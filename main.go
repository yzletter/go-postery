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
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/service/ratelimit"

	infraMySQL "github.com/yzletter/go-postery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/infra/redis"
)

func main() {
	// Infra 层
	infraMySQL.Init("./conf", "db", viper.YAML, "./logs") // 注册 MySQL
	infraRedis.Init("./conf", "redis", viper.YAML)        // 注册 Redis
	slog.InitSlog("./logs/go_postery.logs")               // 初始化 slog
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

	JwtSvc := service.NewJwtService("123456")                                               // 注册 JwtService
	MetricSvc := service.NewMetricService()                                                 // 注册 MetricService
	RateLimitSvc := ratelimit.NewRateLimitService(infraRedis.GetRedis(), time.Minute, 1000) // 注册 RateLimitService
	AuthSvc := service.NewAuthService(infraRedis.GetRedis(), JwtSvc, UserSvc)               // 注册 AuthService

	// Handler 层
	UserHdl := handler.NewUserHandler(AuthSvc, JwtSvc, UserSvc)           // 注册 UserHandler
	PostHdl := handler.NewPostHandler(PostSvc, UserSvc, TagSvc)           // 注册 PostHandler
	CommentHdl := handler.NewCommentHandler(CommentSvc, UserSvc, PostSvc) // 注册 CommentHandler
	FollowHdl := handler.NewFollowHandler(FollowSvc, UserSvc)

	//TagHdl := handler.NewTagHandler(TagSvc)                               // 注册 TagHandler

	// 中间件层
	AuthRequiredMdl := middleware.AuthRequiredMiddleware(AuthSvc) // AuthRequiredMdl 强制登录
	AuthOptionalMdl := middleware.AuthOptionalMiddleware(AuthSvc) // AuthOptionalMdl 非强制要求登录
	AuthAdminMdl := middleware.AuthAdminMiddleware(AuthSvc)       // AuthAdminMdl 要求管理员身份
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
	engine.GET("/posts_tag", PostHdl.ListByTag)      // 根据标签获取帖子列表
	engine.GET("/posts/:pid", PostHdl.Detail)        // 获取帖子详情
	engine.GET("/posts_uid/:uid", PostHdl.ListByUid) // 获取目标用户发布的帖子
	// 强制登录
	engine.POST("/posts/new", AuthRequiredMdl, PostHdl.Create)         // 创建帖子
	engine.GET("/posts/delete/:id", AuthRequiredMdl, PostHdl.Delete)   // 删除帖子
	engine.POST("/posts/update", AuthRequiredMdl, PostHdl.Update)      // 修改帖子
	engine.GET("/posts/like/:id", AuthRequiredMdl, PostHdl.Like)       // 点赞
	engine.GET("/posts/dislike/:id", AuthRequiredMdl, PostHdl.Dislike) // 取消点赞
	engine.GET("/posts/iflike/:id", AuthRequiredMdl, PostHdl.IfLike)   // 取消点赞
	// 非强制要求登录
	engine.GET("/posts/belong", AuthOptionalMdl, PostHdl.Belong) // 查询帖子是否归属当前登录用户

	// 评论模块
	engine.GET("/comment/list/:post_id", CommentHdl.List) // 列出评论
	// 强制登录
	engine.POST("/comment/new", AuthRequiredMdl, CommentHdl.Create)             // 创建评论
	engine.GET("/comment/delete/:pid/:cid", AuthRequiredMdl, CommentHdl.Delete) // 删除评论
	engine.GET("/comment/belong", AuthRequiredMdl, CommentHdl.Belong)           // 删除评论

	// 标签模块
	//engine.POST("/tag/new", AuthRequiredMdl, TagHdl.Create) // 创建标签

	// 关注模块
	engine.GET("/follow/:id", AuthRequiredMdl, FollowHdl.Follow)       // 关注
	engine.GET("/disfollow/:id", AuthRequiredMdl, FollowHdl.DisFollow) // 取消关注
	engine.GET("/iffollow/:id", AuthRequiredMdl, FollowHdl.IfFollow)   // 判断关注关系 0 表示 互不关注 1 表示关注了对方 2 表示对方关注了自己 3 表示互相关注
	engine.GET("/followers", AuthRequiredMdl, FollowHdl.ListFollowers) // 返回粉丝列表
	engine.GET("/followees", AuthRequiredMdl, FollowHdl.ListFollowees) // 返回关注列表

	// 管理员模块
	engine.GET("/admin", AuthRequiredMdl, AuthAdminMdl) // 返回关注列表

	if err := engine.Run("localhost:8765"); err != nil {
		panic(err)
	}
}
