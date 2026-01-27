package main

import (
	"log/slog"
	"net"

	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	"github.com/yzletter/go-postery/search/conf"
	"github.com/yzletter/go-postery/search/infra/kafka"
	infraSlog "github.com/yzletter/go-postery/search/infra/slog"
	"github.com/yzletter/go-postery/search/infra/snowflake"
	"github.com/yzletter/go-postery/search/infra/tokenizer"
	"github.com/yzletter/go-postery/search/service"
	"google.golang.org/grpc"
)

func main() {
	// Infra
	infraSlog.InitSlog(conf.LogFilePath)                // 初始化 slog
	Tokenizer := tokenizer.NewJiebaTokenizer()          // 初始化分词器
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0) // 初始化 雪花算法
	KafkaConsumer := kafka.InitConsumer([]string{conf.KafkaEndpoint}, conf.KafkaTopic, conf.KafkaGroup)

	// Service 层
	SearchService := service.NewSearchService(KafkaConsumer, Tokenizer, IDGenerator)

	// 监听本地端口
	lis, err := net.Listen("tcp", "localhost:"+conf.Port)
	if err != nil {
		panic(err)
	}

	// 启动 grpc 服务
	server := grpc.NewServer()
	search_grpc.RegisterSearchServiceServer(server, SearchService) // 注册服务
	if err := server.Serve(lis); err != nil {
		slog.Error("Search grpc Server Start Failed", "error", err)
		panic(err)
	}
}
