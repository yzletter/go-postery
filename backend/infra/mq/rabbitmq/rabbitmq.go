package infra

import (
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yzletter/go-postery/backend/conf"
)

var globalConn *amqp.Connection

func Init(config conf.RabbitMQConfig) *amqp.Connection {
	mqURL := fmt.Sprintf("amqp://%s:%s@%s/", config.User, config.Password, config.Addr)
	conn, err := amqp.Dial(mqURL)
	if err != nil {
		slog.Error("Init RabbitMQ Connection Failed", "error", err)
	}
	slog.Info("Init RabbitMQ Connection Success")
	globalConn = conn
	return globalConn
}

func Close() {
	if globalConn == nil {
		return
	}

	if err := globalConn.Close(); err != nil {
		slog.Info("Close RabbitMQ Connection Failed", "error", err)
		return
	}

	slog.Info("Close RabbitMQ Success")
}
