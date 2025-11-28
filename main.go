package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	database "github.com/yzletter/go-postery/database/gorm"
	"github.com/yzletter/go-postery/database/redis"
	handler "github.com/yzletter/go-postery/handler/gin"
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

	// 全局中间件
	engine.Use(handler.MetricHandler) // Prometheus 监控中间件

	// 定义路由

	engine.GET("/metrics", func(ctx *gin.Context) { // Prometheus 访问的接口
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request) // 固定写法
	})

	// 用户模块
	engine.POST("/register/submit", handler.RegisterHandlerFunc)                               // 用户注册
	engine.POST("/login/submit", handler.LoginHandlerFunc)                                     // 用户登录
	engine.GET("/logout", handler.LogoutHandlerFunc)                                           // 用户退出
	engine.POST("/modify_pass/submit", handler.AuthHandlerFunc, handler.ModifyPassHandlerFunc) // 修改密码

	// 帖子模块
	engine.GET("/posts", handler.GetPostsHandler)                                       // 获取帖子列表
	engine.GET("/posts/:pid", handler.GetPostDetailHandler)                             // 获取帖子详情
	engine.POST("/posts/new", handler.AuthHandlerFunc, handler.CreateNewPostHandler)    // 创建帖子
	engine.GET("/posts/delete/:id", handler.AuthHandlerFunc, handler.DeletePostHandler) // 删除帖子
	engine.POST("/posts/update", handler.AuthHandlerFunc, handler.UpdatePostHandler)    // 修改帖子
	engine.GET("/posts/belong", handler.PostBelongHandler)                              // 查询帖子是否归属当前登录用户

	if err := engine.Run("localhost:8080"); err != nil {
		panic(err)
	}
}
