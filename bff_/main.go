package main

import (
	"syscall"

	infraQdarant "github.com/yzletter/go-postery/agent/infra/qdrant"
	handler2 "github.com/yzletter/go-postery/bff/handler"
	infraRabbitMQ "github.com/yzletter/go-postery/bff/infra/rabbitmq"
	infraRedis "github.com/yzletter/go-postery/bff/infra/redis"
	"github.com/yzletter/go-postery/bff_/infra/crontab"
	"github.com/yzletter/go-postery/bff_/infra/graceful_stop"
	"github.com/yzletter/go-postery/bff_/service"
	infraRocketMQ "github.com/yzletter/go-postery/lottery/infra/rocketmq"
	infraKafka "github.com/yzletter/go-postery/outbox/infra/kafka"
	"github.com/yzletter/go-postery/outbox/infra/mysql"
)

func main() {
	// Service 层

	// Handler 层

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

	if err := engine.Run("localhost:8765"); err != nil {
		panic(err)
	}
}
