package dao

import (
	"context"

	"github.com/yzletter/go-postery/dto/session"
	"github.com/yzletter/go-postery/model"
)

// 定义 DAO 层所有接口

type MessageDAO interface {
	Create(ctx context.Context, message *model.Message) error
	GetByIDAndTargetID(ctx context.Context, id, targetID int64) ([]*model.Message, error)
	GetByPage(ctx context.Context, id int64, targetID int64, pageNo, pageSize int) (int64, []*model.Message, error)
}

type SessionDAO interface {
	Create(ctx context.Context, session *model.Session) error
	GetByUid(ctx context.Context, uid int64) ([]*model.Session, error)
	GetByUidAndTargetID(ctx context.Context, uid, targetID int64) (*model.Session, error)
	GetByID(ctx context.Context, uid, sid int64) (*model.Session, error)
	Delete(ctx context.Context, uid, sid int64) error
	UpdateUnread(ctx context.Context, uid int64, sid int64, updates session.UpdateUnreadRequest) error
	ClearUnread(ctx context.Context, uid int64, sid int64) error
}
