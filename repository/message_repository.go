package repository

import (
	"context"

	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
)

type messageRepository struct {
	dao   dao.MessageDAO
	cache cache.MessageCache
}

func NewMessageRepository(dao dao.MessageDAO, cache cache.MessageCache) MessageRepository {
	return &messageRepository{dao: dao, cache: cache}
}

func (repo *messageRepository) Create(ctx context.Context, message *model.Message) error {
	err := repo.dao.Create(ctx, message)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *messageRepository) GetByIDAndTargetID(ctx context.Context, id, targetID int64) ([]*model.Message, error) {
	var empty []*model.Message
	messages, err := repo.dao.GetByIDAndTargetID(ctx, id, targetID)
	if err != nil {
		return empty, toRepositoryErr(err)
	}

	return messages, nil
}

func (repo *messageRepository) GetByPage(ctx context.Context, id int64, targetID int64, pageNo, pageSize int) (int, []*model.Message, error) {
	var empty []*model.Message
	total, messages, err := repo.dao.GetByPage(ctx, id, targetID, pageNo, pageSize)
	if err != nil {
		return 0, empty, toRepositoryErr(err)
	}

	return int(total), messages, nil
}
