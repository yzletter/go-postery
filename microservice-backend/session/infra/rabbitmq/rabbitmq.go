package infra

import (
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yzletter/go-postery/microservice-backend/session/config"
)

var globalConn *amqp.Connection

func Init(config config.RabbitMQConfig) *amqp.Connection {
	mqUrl := fmt.Sprintf("amqp://%s:%s@%s/", config.User, config.Password, config.Addr)
	conn, err := amqp.Dial(mqUrl)
	if err != nil {
		slog.Error("初始化 RabbitMQ Connection 失败 ...", "error", err)
	}
	slog.Info("初始化 RabbitMQ Connection 成功 ...")
	globalConn = conn
	return globalConn
}

// Close 关闭 MySQL 连接
func Close() {
	if globalConn != nil {
		err := globalConn.Close()
		if err != nil {
			slog.Info("关闭 RabbitMQ Connection 失败 ...")
		}

		slog.Info("关闭 RabbitMQ 成功 ...")
		return
	}
}
