package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yzletter/go-postery/handler"

	infraMySQL "github.com/yzletter/go-postery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/infra/redis"

	"github.com/yzletter/go-postery/middleware"
	"github.com/yzletter/go-postery/repository/gorm"
	"github.com/yzletter/go-postery/repository/redis"

	postRepository "github.com/yzletter/go-postery/repository/post"
	userRepository "github.com/yzletter/go-postery/repository/user"

	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/crontab"
	"github.com/yzletter/go-postery/utils/smooth"
)

func main() {
	// 初始化
	SlogConfPath := "./log/go_postery.log"
	utils.InitSlog(SlogConfPath) // 初始化 slog

	crontab.InitCrontab()   // 初始化 定时任务
	smooth.InitSmoothExit() // 初始化 优雅退出

	database.ConnectToMySQL("./conf", "db", utils.YAML, "./log") // 初始化 MySQL 数据库
	redis.ConnectToRedis("./conf", "redis", utils.YAML)          // 初始化 Redis 数据库

	// 初始化 gin
	engine := gin.Default()

	// 配置跨域
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // 允许域名跨域
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Infra 层
	infraMySQL.Init("./conf", "db", utils.YAML, "./log") // 注册 MySQL
	infraRedis.Init("./conf", "redis", utils.YAML)       // 注册 Redis

	// Repository 层
	UserRepo := userRepository.NewGormUserRepository(infraMySQL.GetDB()) // 注册 UserRepository
	PostRepo := postRepository.NewGormPostRepository(infraMySQL.GetDB()) // 注册 PostRepository

	// Service 层
	JwtService := service.NewJwtService("123456")                            // 注册 JwtService
	UserService := service.NewUserService(UserRepo)                          // 注册 UserService
	PostService := service.NewPostService(PostRepo)                          // 注册 PostService
	MetricService := service.NewMetricService()                              // 注册 MetricService
	AuthService := service.NewAuthService(infraRedis.GetRedis(), JwtService) // 注册 AuthService

	// Handler 层
	// todo 会换成 infra
	UserHdl := handler.NewUserHandler(infraRedis.GetRedis(), JwtService, UserService) // 注册 UserHandler
	PostHdl := handler.NewPostHandler(PostService)                                    // 注册 PostHandler

	// 中间件层
	AuthRequiredMdl := middleware.AuthRequiredMiddleware(AuthService) // AuthRequiredMdl 强制登录
	AuthOptionalMdl := middleware.AuthOptionalMiddleware(AuthService) // AuthOptionalMdl 非强制要求登录
	MetricMdl := middleware.MetricMiddleware(MetricService)           // MetricMdl 用于 Prometheus 监控中间件

	// 全局中间件
	engine.Use(MetricMdl) // Prometheus 监控中间件

	// 定义路由
	engine.GET("/metrics", func(ctx *gin.Context) { // Prometheus 访问的接口
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request) // 固定写法
	})

	// 用户模块
	engine.POST("/register/submit", UserHdl.Register) // 用户注册
	engine.POST("/login/submit", UserHdl.Login)       // 用户登录
	engine.GET("/logout", UserHdl.Logout)             // 用户退出
	// 强制登录
	engine.POST("/modify_pass/submit", AuthRequiredMdl, UserHdl.ModifyPass) // 修改密码

	// 帖子模块
	engine.GET("/posts", PostHdl.List)        // 获取帖子列表
	engine.GET("/posts/:pid", PostHdl.Detail) // 获取帖子详情
	// 强制登录
	engine.POST("/posts/new", AuthRequiredMdl, PostHdl.Create)       // 创建帖子
	engine.GET("/posts/delete/:id", AuthRequiredMdl, PostHdl.Delete) // 删除帖子
	engine.POST("/posts/update", AuthRequiredMdl, PostHdl.Update)    // 修改帖子
	// 非强制要求登录
	engine.GET("/posts/belong", AuthOptionalMdl, PostHdl.Belong) // 查询帖子是否归属当前登录用户

	if err := engine.Run("localhost:8080"); err != nil {
		panic(err)
	}
}
