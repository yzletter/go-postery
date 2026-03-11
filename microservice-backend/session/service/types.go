package service

import (
	"context"

	model2 "github.com/yzletter/go-postery/microservice-backend/session/model"
)

type SessionService interface {
	ListByUID(ctx context.Context, userID int64) ([]*model2.Session, error)
	GetSession(ctx context.Context, userID int64, targetID int64) (*model2.Session, error)
	GetHistoryMessagesByPage(ctx context.Context, userID int64, targetID int64, pageNo int, pageSize int) (int64, []*model2.Message, error)
	Delete(ctx context.Context, userID int64, sessionID int64) error
	UpdateUnread(ctx context.Context, userID int64, sessionID int64, updates model2.UpdateUnread) error
	ClearUnread(ctx context.Context, userID int64, sessionID int64) error
	CreateMessage(ctx context.Context, message *model2.Message) (*model2.Message, error)
	StartSessionRegisterConsumer(ctx context.Context)
}
