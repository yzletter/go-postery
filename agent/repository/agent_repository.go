package repository

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
	"github.com/yzletter/go-postery/agent/model"
	"github.com/yzletter/go-postery/agent/repository/dao"
)

type agentRepository struct {
	dao dao.AgentDAO
}

func NewAgentRepository(dao dao.AgentDAO) AgentRepository {
	return &agentRepository{dao: dao}
}

func (repo *agentRepository) Retrieve(ctx context.Context, query string, scoreThreshold float64, limit int) ([]string, error) {
	texts, err := repo.dao.Retrieve(ctx, query, scoreThreshold, limit)
	if err != nil {
		return []string{}, toRepositoryErr(err)
	}

	return texts, nil
}

func (repo *agentRepository) CreateChunksWithOutbox(ctx context.Context, chunkModels []*model.Chunk, event *model.Event) error {
	err := repo.dao.CreateChunksWithOutbox(ctx, chunkModels, event)
	if err != nil {
		return toRepositoryErr(err)
	}

	return nil
}

func (repo *agentRepository) UpsertVectorPoints(ctx context.Context, points []*qdrant.PointStruct) error {
	err := repo.dao.UpsertVectorPoints(ctx, points)
	if err != nil {
		return toRepositoryErr(err)
	}

	return nil
}

func (repo *agentRepository) GetChunksByBatchID(ctx context.Context, BatchID int64) ([]*model.Chunk, error) {
	chunks, err := repo.dao.GetChunksByBatchID(ctx, BatchID)
	if err != nil {
		return []*model.Chunk{}, toRepositoryErr(err)
	}

	return chunks, nil
}
