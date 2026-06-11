package repository

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
	model2 "github.com/yzletter/go-postery/backend/micro/agent/model"
)

type AgentRepository interface {
	Retrieve(ctx context.Context, query string, scoreThreshold float64, limit int) ([]string, error)
	CreateChunksWithOutbox(ctx context.Context, chunkModels []*model2.Chunk, event *model2.Event) error
	UpsertVectorPoints(ctx context.Context, points []*qdrant.PointStruct) error
	GetChunksByBatchID(ctx context.Context, BatchID int64) ([]*model2.Chunk, error)
}
