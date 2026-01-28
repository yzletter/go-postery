package service

import (
	"context"

	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
)

type SessionService interface {
	ListByUID(context.Context, *session_grpc.UserID) (*session_grpc.Sessions, error)
	GetSession(context.Context, *session_grpc.BothUserID) (*session_grpc.Session, error)
	GetHistoryMessagesByPage(context.Context, *session_grpc.GetHistoryMessagesByPageRequest) (*session_grpc.GetHistoryMessagesByPageResponse, error)
	Delete(context.Context, *session_grpc.DeleteRequest) (*session_grpc.SessionEmptyResponse, error)
	UpdateUnread(context.Context, *session_grpc.UpdateUnreadRequest) (*session_grpc.SessionEmptyResponse, error)
	ClearUnread(context.Context, *session_grpc.ClearUnreadRequest) (*session_grpc.SessionEmptyResponse, error)
	CreateMessage(context.Context, *session_grpc.Message) (*session_grpc.SessionEmptyResponse, error)
	StartSessionRegisterConsumer(ctx context.Context)
	session_grpc.UnsafeSessionServiceServer
}
