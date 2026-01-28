package main

import (
	"context"
	"log/slog"
	"net"

	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"github.com/yzletter/go-postery/session/conf"
	infraKafka "github.com/yzletter/go-postery/session/infra/kafka"
	infraMySQL "github.com/yzletter/go-postery/session/infra/mysql"
	infraRabbitMQ "github.com/yzletter/go-postery/session/infra/rabbitmq"
	infraSlog "github.com/yzletter/go-postery/session/infra/slog"
	"github.com/yzletter/go-postery/session/infra/snowflake"
	"github.com/yzletter/go-postery/session/infra/viper"
	"github.com/yzletter/go-postery/session/repository"
	"github.com/yzletter/go-postery/session/repository/dao"
	"github.com/yzletter/go-postery/session/service"
	"google.golang.org/grpc"
)

func main() {
	// Infra 层
	infraSlog.InitSlog(conf.LogFilePath)                                                                // 初始化 slog
	RabbitMQ := infraRabbitMQ.Init("./conf", "mq", viper.YAML)                                          // 初始化 RabbitMQ
	MySQLGormDB := infraMySQL.Init("./conf", "db", viper.YAML, "./logs")                                // 初始化 MySQL
	SessionKafkaConsumer := infraKafka.InitConsumer([]string{conf.KafkaEndpoint}, "session", "session") // 初始化 Session 模块 Kafka 消费方
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)                                                 // 初始化 雪花算法
	// DAO 层
	MessageDAO := dao.NewMessageDAO(MySQLGormDB)
	SessionDAO := dao.NewSessionDAO(MySQLGormDB)
	// Repository 层
	MessageRepo := repository.NewMessageRepository(MessageDAO) // 注册 MessageRepository
	SessionRepo := repository.NewSessionRepository(SessionDAO) // 注册 SessionRepository
	// Service 层
	SessionService := service.NewSessionService(SessionRepo, MessageRepo, RabbitMQ, SessionKafkaConsumer, IDGenerator) // 注册 SessionService
	ctx := context.Background()
	go SessionService.StartSessionRegisterConsumer(ctx) // 开启协程注册新用户聊天功能

	// 监听本地端口
	lis, err := net.Listen("tcp", "localhost:"+conf.Port)
	if err != nil {
		panic(err)
	}

	// 启动 grpc 服务
	server := grpc.NewServer()
	session_grpc.RegisterSessionServiceServer(server, SessionService) // 注册服务
	if err := server.Serve(lis); err != nil {
		slog.Error("Code grpc Server Start Failed", "error", err)
		panic(err)
	}
}
