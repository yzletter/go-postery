package service

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	database "github.com/yzletter/go-postery/database/gorm"
	"github.com/yzletter/go-postery/database/redis"
)

func InitSmoothExit() {
	var listen func()
	// 监听函数
	listen = func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM) // 注册信号 2 和 15
		s := <-ch                                          // 阻塞, 直到信号到来
		slog.Info("signal " + s.String() + " has come, start exiting ...")

		// 退出前具体要做的工作
		database.CloseConnection() // 这里以关闭数据库连接为例
		redis.CloseConnection()

		slog.Info("all task has finished")

		// 退出所有进程
		os.Exit(0)
	}
	go listen() // 开始监听信号
}
