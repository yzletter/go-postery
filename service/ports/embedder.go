package ports

import (
	"context"
	"errors"
)

type Embedder interface {
	Embedding(ctx context.Context, text []string) ([][]float64, error)
}

var (
	ErrEmbeddingFailed = errors.New("向量计算失败")
)
