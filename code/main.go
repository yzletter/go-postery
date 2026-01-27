package main

import (
	"log/slog"
	"net"
	"os"

	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/code/conf"
	"github.com/yzletter/go-postery/code/infra/email"
	infraRedis "github.com/yzletter/go-postery/code/infra/redis"
	infraSlog "github.com/yzletter/go-postery/code/infra/slog"
	"github.com/yzletter/go-postery/code/infra/sms"
	"github.com/yzletter/go-postery/code/infra/viper"
	"github.com/yzletter/go-postery/code/repository"
	"github.com/yzletter/go-postery/code/repository/cache"
	"github.com/yzletter/go-postery/code/service"
	"google.golang.org/grpc"
)

func main() {
	// Infra
	infraSlog.InitSlog(conf.LogFilePath)                                                                                    // 初始化 slog
	RedisClient := infraRedis.Init("./conf", "cache", viper.YAML)                                                           // 初始化 Redis
	SmsClient := sms.NewAliyunSmsClient(os.Getenv(conf.AliyunAccessTokenKeyID), os.Getenv(conf.AliyunAccessTokenKeySecret)) // 初始化 短信服务商
	EmailManager := email.NewEmailManager(conf.EmailFrom, os.Getenv(conf.EmailAuthCode), conf.EmailSubject, conf.EmailExpireMin, conf.AppName, conf.EmailYear, conf.Address)
	// Cache
	CodeCache := cache.NewCodeCache(RedisClient)
	// Repository
	CodeRepository := repository.NewCodeRepository(CodeCache)
	// Service
	CodeService := service.NewCodeService(CodeRepository, EmailManager, SmsClient)

	// 监听本地端口
	lis, err := net.Listen("tcp", "localhost:"+conf.Port)
	if err != nil {
		panic(err)
	}

	// 启动 grpc 服务
	server := grpc.NewServer()
	code_grpc.RegisterCodeServiceServer(server, CodeService) // 注册服务
	if err := server.Serve(lis); err != nil {
		slog.Error("Code grpc Server Start Failed", "error", err)
		panic(err)
	}
}
