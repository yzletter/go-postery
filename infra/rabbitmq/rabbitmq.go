package infra

import (
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yzletter/go-postery/infra/viper"
)

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

var global *RabbitMQ

func Init(confDir, confFileName, confFileType string) *RabbitMQ {
	// 初始化一个 Viper 进行配置读取
	vip := viper.InitViper(confDir, confFileName, confFileType)
	user := vip.GetString("RabbitMQ.user")
	password := vip.GetString("RabbitMQ.password")
	host := vip.GetString("RabbitMQ.host")
	port := vip.GetString("RabbitMQ.port")

	mqUrl := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
	conn, err := amqp.Dial(mqUrl)
	if err != nil {
		slog.Error("初始化 RabbitMQ Connection 失败 ...", "error", err)
	}

	// 创建 Channel
	ch, err := conn.Channel()
	if err != nil {
		slog.Error("初始化 RabbitMQ Channel 失败 ...", "error", err)
	}

	mq := &RabbitMQ{
		conn: conn,
		ch:   ch,
	}

	global = mq
	slog.Error("初始化 RabbitMQ 成功 ...", "error", err)
	return mq
}

// Close 关闭 MySQL 连接
func Close() {
	if global != nil {

		err := global.ch.Close()
		if err != nil {
			slog.Info("关闭 RabbitMQ Channel 失败 ...")
		}

		err = global.conn.Close()
		if err != nil {
			slog.Info("关闭 RabbitMQ Connection 失败 ...")
		}

		slog.Info("关闭 RabbitMQ 成功 ...")
		return
	}
}
