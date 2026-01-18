package llm

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
)

var (
	embedder *ark.Embedder
	once     sync.Once
)

func NewArkEmbedder(ctx context.Context, model, APIKey string) *ark.Embedder {
	once.Do(func() {
		var err error
		timeout := 3 * time.Second
		retryTimes := 3
		apiType := ark.APITypeMultiModal
		embedder, err = ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
			Timeout:    &timeout,
			RetryTimes: &retryTimes,
			APIKey:     APIKey,
			Model:      model,
			APIType:    &apiType,
		})
		if err != nil {
			slog.Error("初始化 ArkEmbedder 失败 ...")
			return
		}
	})
	slog.Error("初始化 ArkEmbedder 成功 ...")
	return embedder
}
