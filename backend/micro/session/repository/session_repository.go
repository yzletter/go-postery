package repository

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/session/domain"
	"github.com/yzletter/go-postery/backend/micro/session/repository/dao"
)

type sessionRepository struct {
	dao dao.SessionDAO
}

func NewSessionRepository(dao dao.SessionDAO) SessionRepository {
	return &sessionRepository{
		dao: dao,
	}
}

func (repo *sessionRepository) Create(ctx context.Context, session ...domain.Session) error {
	models := domain.ToModelSessions(session...)
	if err := repo.dao.Create(ctx, models...); err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *sessionRepository) Recover(ctx context.Context, uid, targetID int64) (domain.Session, error) {
	session, err := repo.dao.Recover(ctx, uid, targetID)
	if err != nil {
		return domain.Session{}, toRepositoryErr(err)
	}
	return domain.ToDomainSession(session), nil
}

func (repo *sessionRepository) GetByUidAndTargetID(ctx context.Context, uid, targetID int64) (domain.Session, error) {
	// 查数据库
	session, err := repo.dao.GetByUidAndTargetID(ctx, uid, targetID)
	if err != nil {
		return domain.Session{}, toRepositoryErr(err)
	}

	return domain.ToDomainSession(session), nil
}

func (repo *sessionRepository) ListByUid(ctx context.Context, uid int64) ([]domain.Session, error) {
	// todo 查缓存

	// 查数据库
	sessions, err := repo.dao.GetByUid(ctx, uid)
	if err != nil {
		return nil, ErrServerInternal
	}

	return domain.ToDomainSessions(sessions), nil
}

func (repo *sessionRepository) GetByID(ctx context.Context, uid, sid int64) (domain.Session, error) {
	session, err := repo.dao.GetByID(ctx, uid, sid)
	if err != nil {
		return domain.Session{}, toRepositoryErr(err)
	}
	return domain.ToDomainSession(session), nil
}

func (repo *sessionRepository) Delete(ctx context.Context, uid, sid int64) error {
	err := repo.dao.Delete(ctx, uid, sid)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *sessionRepository) UpdateUnread(ctx context.Context, uid int64, sid int64, updates domain.UpdateUnread) error {
	err := repo.dao.UpdateUnread(ctx, uid, sid, domain.ToModelUpdateUnread(updates))
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
