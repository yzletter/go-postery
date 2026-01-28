package repository

import (
	"context"

	"github.com/yzletter/go-postery/session/model"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	ListByUid(ctx context.Context, uid int64) ([]*model.Session, error)
	GetByUidAndTargetID(ctx context.Context, uid, targetID int64) (*model.Session, error)
	GetByID(ctx context.Context, uid, sid int64) (*model.Session, error)
	Delete(ctx context.Context, uid, sid int64) error
	UpdateUnread(ctx context.Context, uid int64, sid int64, updates model.UpdateUnread) error
	ClearUnread(ctx context.Context, uid int64, sid int64) error
}

type MessageRepository interface {
	Create(ctx context.Context, message *model.Message) error
	GetByIDAndTargetID(ctx context.Context, id, targetID int64) ([]*model.Message, error)
	GetByPage(ctx context.Context, id int64, targetID int64, pageNo, pageSize int) (int, []*model.Message, error)
}
