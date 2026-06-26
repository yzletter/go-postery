package repository

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/session/domain"
	"github.com/yzletter/go-postery/backend/micro/session/repository/dao"
)

type messageRepository struct {
	dao dao.MessageDAO
}

func NewMessageRepository(dao dao.MessageDAO) MessageRepository {
	return &messageRepository{dao: dao}
}

func (repo *messageRepository) Create(ctx context.Context, message domain.Message) error {
	m := domain.ToModelMessage(message)
	err := repo.dao.Create(ctx, &m)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *messageRepository) GetByIDAndTargetID(ctx context.Context, id, targetID int64) ([]domain.Message, error) {
	messages, err := repo.dao.GetByIDAndTargetID(ctx, id, targetID)
	if err != nil {
		return nil, toRepositoryErr(err)
	}

	return domain.ToDomainMessages(messages), nil
}

func (repo *messageRepository) GetByPage(ctx context.Context, id int64, targetID int64, pageNo, pageSize int) (int, []domain.Message, error) {
	total, messages, err := repo.dao.GetByPage(ctx, id, targetID, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	return int(total), domain.ToDomainMessages(messages), nil
}
