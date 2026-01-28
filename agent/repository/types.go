package repository

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
	"github.com/yzletter/go-postery/agent/model"
)

type AgentRepository interface {
	Retrieve(ctx context.Context, query string, scoreThreshold float64, limit int) ([]string, error)
	CreateChunksWithOutbox(ctx context.Context, chunkModels []*model.Chunk, event *model.Event) error
	UpsertVectorPoints(ctx context.Context, points []*qdrant.PointStruct) error
	GetChunksByBatchID(ctx context.Context, BatchID int64) ([]*model.Chunk, error)
}
