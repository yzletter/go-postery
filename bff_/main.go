package main

import (
	"syscall"

	infraQdarant "github.com/yzletter/go-postery/agent/infra/qdrant"
	handler2 "github.com/yzletter/go-postery/bff/handler"
	infraRedis "github.com/yzletter/go-postery/bff/infra/redis"
	"github.com/yzletter/go-postery/bff/infra/viper"
	"github.com/yzletter/go-postery/bff_/conf"
	"github.com/yzletter/go-postery/bff_/handler"
	"github.com/yzletter/go-postery/bff_/infra/crontab"
	"github.com/yzletter/go-postery/bff_/infra/graceful_stop"
	infraRabbitMQ "github.com/yzletter/go-postery/bff_/infra/rabbitmq"
	"github.com/yzletter/go-postery/bff_/service"
	infraRocketMQ "github.com/yzletter/go-postery/lottery/infra/rocketmq"
	infraKafka "github.com/yzletter/go-postery/outbox/infra/kafka"
	"github.com/yzletter/go-postery/outbox/infra/mysql"
)

func main() {
	// Infra 层
	RabbitMQ := infraRabbitMQ.Init("./conf", "mq", viper.YAML) // 初始化 RabbitMQ

	// GRPC Service 层
	SessionServiceClient := service.NewSessionService("localhost:" + conf.SessionPort)

	// Service 层
	WebsocketSvc := service.NewWebsocketService(SessionServiceClient, RabbitMQ) // 注册 WebsocketService

	// Handler 层
	SessionHdl := handler.NewSessionHandler(SessionServiceClient) // 注册 SessionHandler
	WebsocketHdl := handler.NewWebsocketHandler(WebsocketSvc)     // 注册 WebsocketHandler

	// 中间件层

	// 初始化 Crontab
	crontab.NewCrontabBuilder().
		AddFuncWithSpec("*/10 * * * *", infra.Ping).
		AddFuncWithSpec("*/10 * * * *", infraRedis.Ping).
		Build()

	// 初始化 GracefulStop
	graceful_stop.NewGracefulStopBuilder().
		NotifySignal(syscall.SIGINT).NotifySignal(syscall.SIGTERM).
		AddFunc(infraRabbitMQ.Close).AddFunc(infraRocketMQ.Close).AddFunc(infraKafka.Close). // 关消息队列
		AddFunc(infra.Close).AddFunc(infraRedis.Close).AddFunc(infraQdarant.Close).          // 关数据库
		Build()

	// 私信模块
	sessions := v1.Group("/sessions")
	sessions.Use(AuthRequiredMdl)
	{
		sessions.GET("", SessionHdl.List)          // GET /api/v1/sessions							获取当前登录用户会话列表
		sessions.DELETE("/:id", SessionHdl.Delete) // DELETE /api/v1/sessions/:id						删除当前会话
	}

	// 即时聊天模块
	im := v1.Group("/ws")
	im.Use(AuthRequiredMdl)
	{
		im.GET("", WebsocketHdl.Connect) // GET /api/v1/ws
	}

	if err := engine.Run("localhost:8765"); err != nil {
		panic(err)
	}
}
