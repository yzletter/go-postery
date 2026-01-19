package llm

import (
	"context"
	"log/slog"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/yzletter/go-postery/service/ports"
)

type ArkEmbedder struct {
	embedder *ark.Embedder
}

func NewArkEmbedder(ctx context.Context, model, APIKey string) *ArkEmbedder {
	var err error
	timeout := 3 * time.Second
	retryTimes := 3
	apiType := ark.APITypeMultiModal

	// 采用火山方舟向量模型
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		Timeout:    &timeout,
		RetryTimes: &retryTimes,
		APIKey:     APIKey,
		Model:      model,
		APIType:    &apiType,
	})
	if err != nil {
		slog.Error("初始化 ArkEmbedder 失败 ...")
		return nil
	}

	slog.Error("初始化 ArkEmbedder 成功 ...")
	return &ArkEmbedder{embedder: embedder}
}

func (e ArkEmbedder) Embedding(ctx context.Context, text []string) ([][]float64, error) {
	embeddings, err := e.embedder.EmbedStrings(ctx, text)
	if err != nil {
		return nil, ports.ErrEmbeddingFailed
	}

	return embeddings, nil
}
