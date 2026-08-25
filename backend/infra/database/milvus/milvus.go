package milvus

import (
	"context"
	"log/slog"
	"sync"

	milvusSDK "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/yzletter/go-postery/backend/conf"
)

var (
	client milvusSDK.Client
	once   sync.Once
)

func NewMilvusClient(ctx context.Context, config conf.MilvusConfig) milvusSDK.Client {
	once.Do(func() {
		var err error
		client, err = milvusSDK.NewClient(ctx, milvusSDK.Config{
			Address: config.Addr,
		})
		if err != nil {
			slog.Error("init milvus failed ...", "error", err.Error())
			return
		}
	})

	return client
}

func Close() {
	_ = client.Close()
}
