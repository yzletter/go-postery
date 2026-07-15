package milvus

import (
	"context"
	"sync"

	milvusSDK "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/yzletter/go-postery/backend/conf"
)

var (
	client milvusSDK.Client
	once   sync.Once
)

func NewMilvusClient(ctx context.Context, config conf.MilvusConfig) milvusSDK.Client {
	var err error
	once.Do(func() {
		client, err = milvusSDK.NewClient(ctx, milvusSDK.Config{
			Address: config.Addr,
		})

		if err != nil {
			return
		}

	})

	return client
}

func Close() {
	_ = client.Close()
}
