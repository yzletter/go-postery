package qdrant

import (
	"log/slog"
	"sync"

	"github.com/qdrant/go-client/qdrant"
	"github.com/yzletter/go-postery/microservice-backend/agent/config"
)

var (
	client *qdrant.Client
	once   sync.Once
)

func Init(config config.QdrantConfig) *qdrant.Client {
	once.Do(func() {
		// 读取 Qdrant 相关配置
		var err error
		client, err = qdrant.NewClient(&qdrant.Config{
			Host: config.Host,
			Port: config.Port,
		})
		if err != nil {
			slog.Info("初始化 Qdrant 失败 ...", "error", err)
		}
	})

	return client
}

func Close() {
	if client != nil {
		err := client.Close()
		if err != nil {
			slog.Info("关闭 Qdrant 失败 ...")
			return
		}
		slog.Info("关闭 Qdrant 成功 ...")
	}
}
