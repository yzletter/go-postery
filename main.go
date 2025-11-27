package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	database "github.com/yzletter/go-postery/database/gorm"
	handler "github.com/yzletter/go-postery/handler/gin"
	"github.com/yzletter/go-postery/utils"
)

func main() {
	// 初始化
	utils.InitSlog("./log/go_postery.log")
	database.ConnectToDB("./conf", "db", utils.YAML, "./log")

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

	// 定义路由

	// 用户模块
	engine.POST("/register/submit", handler.RegisterHandlerFunc)                               // 用户注册
	engine.POST("/login/submit", handler.LoginHandlerFunc)                                     // 用户登录
	engine.GET("/logout", handler.LogoutHandlerFunc)                                           // 用户退出
	engine.POST("/modify_pass/submit", handler.AuthHandlerFunc, handler.ModifyPassHandlerFunc) // 修改密码

	// 帖子模块
	engine.GET("/posts", handler.GetPostsHandler)                                    // 获取帖子列表
	engine.GET("/posts/:pid", handler.GetPostDetailHandler)                          // 获取帖子详情
	engine.POST("/posts/new", handler.AuthHandlerFunc, handler.CreateNewPostHandler) // 创建帖子

	if err := engine.Run("localhost:8080"); err != nil {
		panic(err)
	}
}
