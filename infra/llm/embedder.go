package llm

import (
	"context"
	"log/slog"
	"math"
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

	// 向量归一化
	for i, vector := range embeddings {
		embeddings[i] = e.NormVector(vector)
	}

	return embeddings, nil
}

func (e ArkEmbedder) NormVector(vec []float64) []float64 {
	// 检查参数
	if vec == nil || len(vec) == 0 {
		return nil
	}

	// 计算模长
	sum := 0.
	for _, degree := range vec {
		sum += degree * degree
	}
	norm := math.Sqrt(sum)

	for i := range vec {
		vec[i] /= norm
	}

	return vec
}

// AvgOfVector 多个向量按位求平均
func (e ArkEmbedder) AvgOfVector(vectors [][]float64) ([]float64, error) {
	n := len(vectors)
	if n == 0 {
		return nil, ports.ErrInvalidEmbeddingParams
	} else if n == 1 {
		// 向量归一化
		return e.NormVector(vectors[0]), nil
	}

	l := len(vectors[0])
	sum := make([]float64, l)
	for i := 0; i < n; i++ {
		if len(vectors[i]) != l {
			return nil, ports.ErrInvalidEmbeddingParams
		}
		for j := 0; j < l; j++ {
			sum[j] += vectors[i][j] //按位求和
		}
	}
	for j := 0; j < l; j++ {
		sum[j] /= float64(n) //按位求平均
	}

	// 向量归一化
	return e.NormVector(sum), nil
}
