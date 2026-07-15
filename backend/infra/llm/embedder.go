package llm

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/yzletter/go-postery/backend/conf"
)

var (
	ErrEmbeddingFailed        = errors.New("embedding failed")
	ErrInvalidEmbeddingParams = errors.New("invalid embedding params")
)

type Embedder interface {
	EmbedStrings(ctx context.Context, text []string, opts ...embedding.Option) ([][]float64, error) // 实现 eino 接口
	MyEmbedStrings(ctx context.Context, text []string) ([][]float64, error)
	NormVector(vec []float64) []float64
	AvgOfVector(vectors [][]float64) ([]float64, error)
}

type ArkEmbedder struct {
	embedder *ark.Embedder
}

func NewArkEmbedder(ctx context.Context, config conf.ArkConfig) *ArkEmbedder {
	timeout := 3 * time.Second
	retryTimes := 3
	apiType := ark.APITypeMultiModal

	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		Timeout:    &timeout,
		RetryTimes: &retryTimes,
		APIKey:     config.APIKey,
		Model:      config.EmbedderModel,
		APIType:    &apiType,
	})
	if err != nil {
		slog.Info("Init ArkEmbedder Failed", "error", err)
		return nil
	}

	slog.Info("Init ArkEmbedder Success")
	return &ArkEmbedder{embedder: embedder}
}

func (e *ArkEmbedder) GetInternal() *ark.Embedder {
	if e == nil {
		return nil
	}
	return e.embedder
}

func (e *ArkEmbedder) EmbedStrings(ctx context.Context, text []string, opts ...embedding.Option) ([][]float64, error) {
	embeddings, err := e.embedder.EmbedStrings(ctx, text, opts...)
	if err != nil {
		return nil, ErrEmbeddingFailed
	}

	for i, vector := range embeddings {
		embeddings[i] = e.NormVector(vector)
	}

	return embeddings, nil
}

func (e *ArkEmbedder) MyEmbedStrings(ctx context.Context, text []string) ([][]float64, error) {
	embeddings, err := e.embedder.EmbedStrings(ctx, text)
	if err != nil {
		return nil, ErrEmbeddingFailed
	}

	for i, vector := range embeddings {
		embeddings[i] = e.NormVector(vector)
	}

	return embeddings, nil
}

func (e *ArkEmbedder) NormVector(vec []float64) []float64 {
	if len(vec) == 0 {
		return nil
	}

	sum := 0.0
	for _, degree := range vec {
		sum += degree * degree
	}
	norm := math.Sqrt(sum)
	if norm == 0 {
		return vec
	}

	for i := range vec {
		vec[i] /= norm
	}

	return vec
}

func (e *ArkEmbedder) AvgOfVector(vectors [][]float64) ([]float64, error) {
	n := len(vectors)
	if n == 0 {
		return nil, ErrInvalidEmbeddingParams
	}
	if n == 1 {
		return e.NormVector(vectors[0]), nil
	}

	l := len(vectors[0])
	sum := make([]float64, l)
	for i := 0; i < n; i++ {
		if len(vectors[i]) != l {
			return nil, ErrInvalidEmbeddingParams
		}
		for j := 0; j < l; j++ {
			sum[j] += vectors[i][j]
		}
	}
	for j := 0; j < l; j++ {
		sum[j] /= float64(n)
	}

	return e.NormVector(sum), nil
}

type QwenEmbedder struct {
	embedder *dashscope.Embedder
}

func NewQwenEmbedder(ctx context.Context, config conf.QwenConfig) *QwenEmbedder {
	timeout := 3 * time.Second
	dimensions := 1024

	embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		APIKey:     config.APIKey,
		Timeout:    timeout,
		Model:      config.EmbedderModel,
		Dimensions: &dimensions,
	})
	if err != nil {
		slog.Info("Init QwenEmbedder Failed", "error", err)
		return nil
	}

	slog.Info("Init QwenEmbedder Success")
	return &QwenEmbedder{embedder: embedder}
}

func (e *QwenEmbedder) GetInternal() *dashscope.Embedder {
	if e == nil {
		return nil
	}
	return e.embedder
}

func (e *QwenEmbedder) EmbedStrings(ctx context.Context, text []string, opts ...embedding.Option) ([][]float64, error) {
	embeddings, err := e.embedder.EmbedStrings(ctx, text, opts...)
	if err != nil {
		return nil, ErrEmbeddingFailed
	}

	for i, vector := range embeddings {
		embeddings[i] = e.NormVector(vector)
	}

	return embeddings, nil
}

func (e *QwenEmbedder) MyEmbedStrings(ctx context.Context, text []string) ([][]float64, error) {
	embeddings, err := e.embedder.EmbedStrings(ctx, text)
	if err != nil {
		return nil, ErrEmbeddingFailed
	}

	for i, vector := range embeddings {
		embeddings[i] = e.NormVector(vector)
	}

	return embeddings, nil
}

func (e *QwenEmbedder) NormVector(vec []float64) []float64 {
	if len(vec) == 0 {
		return nil
	}

	sum := 0.0
	for _, degree := range vec {
		sum += degree * degree
	}
	norm := math.Sqrt(sum)
	if norm == 0 {
		return vec
	}

	for i := range vec {
		vec[i] /= norm
	}

	return vec
}

func (e *QwenEmbedder) AvgOfVector(vectors [][]float64) ([]float64, error) {
	n := len(vectors)
	if n == 0 {
		return nil, ErrInvalidEmbeddingParams
	}
	if n == 1 {
		return e.NormVector(vectors[0]), nil
	}

	l := len(vectors[0])
	sum := make([]float64, l)
	for i := 0; i < n; i++ {
		if len(vectors[i]) != l {
			return nil, ErrInvalidEmbeddingParams
		}
		for j := 0; j < l; j++ {
			sum[j] += vectors[i][j]
		}
	}
	for j := 0; j < l; j++ {
		sum[j] /= float64(n)
	}

	return e.NormVector(sum), nil
}
