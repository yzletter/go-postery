package repository

import (
	"context"

	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
)

type sessionRepository struct {
	dao   dao.SessionDAO
	cache cache.SessionCache
}

func NewSessionRepository(dao dao.SessionDAO, cache cache.SessionCache) SessionRepository {
	return &sessionRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *sessionRepository) ListByUid(ctx context.Context, uid int64) ([]*model.Session, error) {
	// todo 查缓存

	// 查数据库
	sessions, err := repo.dao.GetByUid(ctx, uid)
	if err != nil {
		return nil, ErrServerInternal
	}

	return sessions, nil
}
