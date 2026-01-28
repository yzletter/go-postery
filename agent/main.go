package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/yzletter/go-postery/agent/conf"
	infraKafka "github.com/yzletter/go-postery/agent/infra/kafka"
	"github.com/yzletter/go-postery/agent/infra/llm"
	infraMySQL "github.com/yzletter/go-postery/agent/infra/mysql"
	infraQdarant "github.com/yzletter/go-postery/agent/infra/qdrant"
	infraSlog "github.com/yzletter/go-postery/agent/infra/slog"
	"github.com/yzletter/go-postery/agent/infra/snowflake"
	"github.com/yzletter/go-postery/agent/infra/viper"
	"github.com/yzletter/go-postery/agent/repository"
	"github.com/yzletter/go-postery/agent/repository/dao"
	"github.com/yzletter/go-postery/agent/service"
	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	"google.golang.org/grpc"
)

func main() {
	// Infra 层
	infraSlog.InitSlog(conf.LogFilePath)                                                                            // 初始化 slog
	MySQLGormDB := infraMySQL.Init("./conf", "db", viper.YAML, "./logs")                                            // 初始化 MySQL
	QdrantClient := infraQdarant.Init("./conf", "db", viper.YAML)                                                   // 初始化 Qdrant
	IDGenerator := snowflake.NewSnowflakeIDGenerator(0)                                                             // 初始化 雪花算法
	ArkEmbedder := llm.NewArkEmbedder(context.Background(), "doubao-embedding-vision-250615", os.Getenv("ARK_KEY")) // 初始化火山引擎向量模型
	ArkChatModel := llm.NewArkModel(context.Background(), "doubao-seed-1-8-251228", os.Getenv("ARK_KEY"))
	QdrantKafkaConsumer := infraKafka.InitConsumer([]string{conf.KafkaEndpoint}, "upsert_qdrant", "agent_qdrant")
	AgentKafkaConsumer := infraKafka.InitConsumer([]string{conf.KafkaEndpoint}, "index_document", "agent_document")

	// DAO 层
	AgentDAO := dao.NewAgentDAO(MySQLGormDB, QdrantClient, ArkEmbedder.GetInternal())
	// Repository 层
	AgentRepo := repository.NewAgentRepository(AgentDAO)
	// Service 层
	AgentService := service.NewAgentService(AgentRepo, AgentKafkaConsumer, QdrantKafkaConsumer, ArkEmbedder, ArkChatModel, IDGenerator)

	ctx := context.Background()
	go AgentService.StartChunkDocConsumer(ctx)     // 开启切分文档协程
	go AgentService.StartUpsertQdrantConsumer(ctx) // 开启向量数据库协程

	// 监听本地端口
	lis, err := net.Listen("tcp", "localhost:"+conf.Port)
	if err != nil {
		panic(err)
	}

	// 启动 grpc 服务
	server := grpc.NewServer()
	agent_grpc.RegisterAgentServiceServer(server, AgentService) // 注册服务
	if err := server.Serve(lis); err != nil {
		slog.Error("Agent grpc Server Start Failed", "error", err)
		panic(err)
	}
}
