package infra

import (
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yzletter/go-postery/infra/viper"
)

var globalConn *amqp.Connection

func Init(confDir, confFileName, confFileType string) *amqp.Connection {
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
