package repository

import (
	"context"

	"github.com/yzletter/go-postery/microservice-backend/session/model"
	"github.com/yzletter/go-postery/microservice-backend/session/repository/dao"
)

type sessionRepository struct {
	dao dao.SessionDAO
}

func NewSessionRepository(dao dao.SessionDAO) SessionRepository {
	return &sessionRepository{
		dao: dao,
	}
}

func (repo *sessionRepository) Create(ctx context.Context, session *model.Session) error {
	err := repo.dao.Create(ctx, session)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *sessionRepository) GetByUidAndTargetID(ctx context.Context, uid, targetID int64) (*model.Session, error) {
	// 查数据库
	session, err := repo.dao.GetByUidAndTargetID(ctx, uid, targetID)
	if err != nil {
		return nil, toRepositoryErr(err)
	}

	return session, nil
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

func (repo *sessionRepository) GetByID(ctx context.Context, uid, sid int64) (*model.Session, error) {
	session, err := repo.dao.GetByID(ctx, uid, sid)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return session, nil
}

func (repo *sessionRepository) Delete(ctx context.Context, uid, sid int64) error {
	err := repo.dao.Delete(ctx, uid, sid)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *sessionRepository) UpdateUnread(ctx context.Context, uid int64, sid int64, updates model.UpdateUnread) error {
	err := repo.dao.UpdateUnread(ctx, uid, sid, updates)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *sessionRepository) ClearUnread(ctx context.Context, uid int64, sid int64) error {
	err := repo.dao.ClearUnread(ctx, uid, sid)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}
