package ports

import (
	"context"
	"errors"
)

type Embedder interface {
	Embedding(ctx context.Context, text []string) ([][]float64, error)
	NormVector(vec []float64) []float64                 // 向量归一化
	AvgOfVector(vectors [][]float64) ([]float64, error) // 多个向量按位求平均
}

var (
	ErrEmbeddingFailed        = errors.New("向量计算失败")
	ErrInvalidEmbeddingParams = errors.New("向量参数非法")
)
