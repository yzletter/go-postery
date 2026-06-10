package qdrant

import (
	"log/slog"
	"sync"

	"github.com/qdrant/go-client/qdrant"
	"github.com/yzletter/go-postery/backend/conf"
)

var (
	client *qdrant.Client
	once   sync.Once
)

func Init(config conf.QdrantConfig) *qdrant.Client {
	once.Do(func() {
		var err error
		client, err = qdrant.NewClient(&qdrant.Config{
			Host: config.Host,
			Port: config.Port,
		})
		if err != nil {
			slog.Info("Init Qdrant Failed", "error", err)
		}
	})

	return client
}

func Close() {
	if client == nil {
		return
	}

	if err := client.Close(); err != nil {
		slog.Info("Close Qdrant Failed", "error", err)
		return
	}
	slog.Info("Close Qdrant Success")
}
