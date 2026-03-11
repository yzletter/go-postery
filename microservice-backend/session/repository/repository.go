package repository

import (
	"context"

	model2 "github.com/yzletter/go-postery/microservice-backend/session/model"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model2.Session) error
	ListByUid(ctx context.Context, uid int64) ([]*model2.Session, error)
	GetByUidAndTargetID(ctx context.Context, uid, targetID int64) (*model2.Session, error)
	GetByID(ctx context.Context, uid, sid int64) (*model2.Session, error)
	Delete(ctx context.Context, uid, sid int64) error
	UpdateUnread(ctx context.Context, uid int64, sid int64, updates model2.UpdateUnread) error
	ClearUnread(ctx context.Context, uid int64, sid int64) error
}

type MessageRepository interface {
	Create(ctx context.Context, message *model2.Message) error
	GetByIDAndTargetID(ctx context.Context, id, targetID int64) ([]*model2.Message, error)
	GetByPage(ctx context.Context, id int64, targetID int64, pageNo, pageSize int) (int, []*model2.Message, error)
}
