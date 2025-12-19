package repository

import (
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
