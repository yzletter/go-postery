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
	database.ConnectToDB("./conf", "db", "yaml", "./log")

	r := gin.Default()

	// 配置跨域
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有域名跨域
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 定义路由
	r.POST("/login", handler.LoginHandlerFunc)           // 用户登录
	r.GET("/logout", handler.LogoutHandlerFunc)          // 用户退出
	r.POST("/modifypass", handler.ModifyPassHandlerFunc) // 修改密码
	if err := r.Run("127.0.0.1:8080"); err != nil {
		panic(err)
	}
}
