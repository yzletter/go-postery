package repository

import (
	"context"

	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository/dao"
)

type agentRepository struct {
	dao dao.AgentDAO
}

func NewAgentRepository(dao dao.AgentDAO) AgentRepository {
	return &agentRepository{dao: dao}
}

func (repo *agentRepository) Retrieve(ctx context.Context, query string, scoreThreshold float64, limit int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (repo *agentRepository) CreateChunks(ctx context.Context, chunkModels []*model.Chunk, event *model.Event) error {
	//TODO implement me
	panic("implement me")
}

func (repo *agentRepository) UpsertVectors(ctx context.Context, chunkModels []*model.Chunk) error {
	err := repo.dao.UpsertVectors(ctx, chunkModels)
	if err != nil {
		return toRepositoryErr(err)
	}

	return nil
}
