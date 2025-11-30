package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yzletter/go-postery/handler"
	"github.com/yzletter/go-postery/handler/auth"
	"github.com/yzletter/go-postery/middleware"
	"github.com/yzletter/go-postery/repository/gorm"
	"github.com/yzletter/go-postery/repository/redis"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/crontab"
	"github.com/yzletter/go-postery/utils/smooth"
)

func main() {
	// 初始化
	SlogConfPath := "./log/go_postery.log"
	utils.InitSlog(SlogConfPath) // 初始化 slog
	crontab.InitCrontab()        // 初始化 定时任务
	smooth.InitSmoothExit()      // 初始化 优雅退出

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

	// Service 层
	JwtService := service.NewJwtService("123456") // 注册 JwtService
	UserService := service.NewUserService(nil)    // 注册 UserService
	PostService := service.NewPostService(nil)    // 注册 PostService

	// Handler 层
	// todo 会换成 infra
	MetricHandler := handler.NewMetricHandler()
	AuthHandler := auth.NewAuthHandler(redis.GoPosteryRedisClient, JwtService)                 // 注册 AuthHandler
	UserHandler := handler.NewUserHandler(redis.GoPosteryRedisClient, JwtService, UserService) // 注册 UserHandler
	PostHandler := handler.NewPostHandler(PostService)

	// 中间件层, 本质为 gin.HandlerFunc
	AuthRequiredMiddleware := middleware.AuthRequiredMiddleware(AuthHandler) // AuthRequiredMiddleware 强制登录
	AuthOptionalMiddleware := middleware.AuthOptionalMiddleware(AuthHandler) // AuthOptionalMiddleware 非强制要求登录
	MetricMiddleware := middleware.MetricMiddleware(MetricHandler)           // MetricMiddleware 用于 Prometheus 监控中间件

	// 全局中间件
	engine.Use(MetricMiddleware) // Prometheus 监控中间件

	// 定义路由
	engine.GET("/metrics", func(ctx *gin.Context) { // Prometheus 访问的接口
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request) // 固定写法
	})

	// 用户模块
	engine.POST("/register/submit", UserHandler.Register)                              // 用户注册
	engine.POST("/login/submit", UserHandler.Login)                                    // 用户登录
	engine.GET("/logout", UserHandler.Logout)                                          // 用户退出
	engine.POST("/modify_pass/submit", AuthRequiredMiddleware, UserHandler.ModifyPass) // 修改密码

	// 帖子模块
	engine.GET("/posts", PostHandler.GetPosts)                                      // 获取帖子列表
	engine.GET("/posts/:pid", PostHandler.GetPostDetail)                            // 获取帖子详情
	engine.POST("/posts/new", AuthRequiredMiddleware, PostHandler.CreateNewPost)    // 创建帖子
	engine.GET("/posts/delete/:id", AuthRequiredMiddleware, PostHandler.DeletePost) // 删除帖子
	engine.POST("/posts/update", AuthRequiredMiddleware, PostHandler.UpdatePost)    // 修改帖子
	engine.GET("/posts/belong", AuthOptionalMiddleware, PostHandler.PostBelong)     // 查询帖子是否归属当前登录用户

	if err := engine.Run("localhost:8080"); err != nil {
		panic(err)
	}
}
