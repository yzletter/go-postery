package grpc

import (
	"context"
	"strings"

	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/micro/session/domain"
	"github.com/yzletter/go-postery/backend/micro/session/grpc/dto"
	"github.com/yzletter/go-postery/backend/micro/session/service"
	"github.com/yzletter/go-postery/backend/utils"
)

type SessionServiceServer struct {
	svc service.SessionService
	session_grpc.UnimplementedSessionServiceServer
}

func NewSessionServiceServer(svc service.SessionService) *SessionServiceServer {
	return &SessionServiceServer{
		svc: svc,
	}
}

// NewConnection 使用 gRPC ctx 维持用户消息队列消费，直到上游 WebSocket 断开。
func (server *SessionServiceServer) NewConnection(ctx context.Context, req *session_grpc.UserID) (*session_grpc.SessionEmptyResponse, error) {
	if req == nil || req.UserID <= 0 {
		return &session_grpc.SessionEmptyResponse{}, errs.ErrInvalidArgument
	}
	if err := server.svc.NewConnection(ctx, req.UserID); err != nil {
		return &session_grpc.SessionEmptyResponse{}, err
	}
	return &session_grpc.SessionEmptyResponse{}, nil
}

// Chat 校验鉴权用户和消息参数后交给 Session Service 处理。
func (server *SessionServiceServer) Chat(ctx context.Context, req *session_grpc.ChatRequest) (*session_grpc.SessionEmptyResponse, error) {
	if req == nil || req.UserID <= 0 || req.Message == nil {
		return &session_grpc.SessionEmptyResponse{}, errs.ErrInvalidArgument
	}
	message := req.Message
	if message.SessionID <= 0 || message.MessageTo <= 0 || strings.TrimSpace(message.Content) == "" {
		return &session_grpc.SessionEmptyResponse{}, errs.ErrInvalidArgument
	}
	if message.MessageFrom != 0 && message.MessageFrom != req.UserID {
		return &session_grpc.SessionEmptyResponse{}, errs.ErrUnauthenticated
	}

	messageDomain := domain.Message{
		SessionID:   message.SessionID,
		SessionType: int(message.SessionType),
		MessageFrom: req.UserID,
		MessageTo:   message.MessageTo,
		Content:     message.Content,
	}
	if err := server.svc.Chat(ctx, req.UserID, messageDomain); err != nil {
		return &session_grpc.SessionEmptyResponse{}, err
	}
	return &session_grpc.SessionEmptyResponse{}, nil
}

func (server *SessionServiceServer) ListByUID(ctx context.Context, id *session_grpc.UserID) (*session_grpc.Sessions, error) {
	sessions, err := server.svc.ListByUID(ctx, id.UserID)
	if err != nil {
		return &session_grpc.Sessions{}, err
	}

	respSessions := make([]*session_grpc.Session, 0, len(sessions))
	for _, session := range sessions {
		respSessions = append(respSessions, dto.ToSession(session))
	}

	return &session_grpc.Sessions{Sessions: respSessions}, nil
}

func (server *SessionServiceServer) GetSession(ctx context.Context, id *session_grpc.BothUserID) (*session_grpc.Session, error) {
	session, err := server.svc.GetSession(ctx, id.UserID, id.TargetID)
	if err != nil {
		return &session_grpc.Session{}, err
	}
	return dto.ToSession(session), nil
}

func (server *SessionServiceServer) GetHistoryMessagesByPage(ctx context.Context, req *session_grpc.GetHistoryMessagesByPageRequest) (*session_grpc.GetHistoryMessagesByPageResponse, error) {
	if req == nil || req.PageNo == 0 || req.PageSize == 0 || req.PageSize > 100 {
		return &session_grpc.GetHistoryMessagesByPageResponse{}, errs.ErrInvalidArgument
	}

	total, messages, err := server.svc.GetHistoryMessagesByPage(ctx, req.UserID, req.TargetID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &session_grpc.GetHistoryMessagesByPageResponse{}, err
	}

	respMessages := make([]*session_grpc.Message, 0, len(messages))
	for _, message := range messages {
		respMessages = append(respMessages, dto.ToMessage(message))
	}

	return &session_grpc.GetHistoryMessagesByPageResponse{
		Count:    uint64(total),
		Messages: respMessages,
	}, nil
}

func (server *SessionServiceServer) Delete(ctx context.Context, req *session_grpc.DeleteRequest) (*session_grpc.SessionEmptyResponse, error) {
	if err := server.svc.Delete(ctx, req.UserID, req.SessionID); err != nil {
		return &session_grpc.SessionEmptyResponse{}, err
	}
	return &session_grpc.SessionEmptyResponse{}, nil
}

func (server *SessionServiceServer) UpdateUnread(ctx context.Context, req *session_grpc.UpdateUnreadRequest) (*session_grpc.SessionEmptyResponse, error) {
	updates := domain.UpdateUnread{
		Updates: domain.Updates{
			LastMessageID:   req.LastMessageID,
			LastMessage:     req.LastMessage,
			LastMessageTime: utils.RPCTimeToGoTime(req.LastMessageTime),
		},
		Delta: int(req.Delta),
	}
	if err := server.svc.UpdateUnread(ctx, req.UserID, req.SessionID, updates); err != nil {
		return &session_grpc.SessionEmptyResponse{}, err
	}
	return &session_grpc.SessionEmptyResponse{}, nil
}

func (server *SessionServiceServer) ClearUnread(ctx context.Context, req *session_grpc.ClearUnreadRequest) (*session_grpc.SessionEmptyResponse, error) {
	if err := server.svc.ClearUnread(ctx, req.UserID, req.SessionID); err != nil {
		return &session_grpc.SessionEmptyResponse{}, err
	}
	return &session_grpc.SessionEmptyResponse{}, nil
}

func (server *SessionServiceServer) CreateMessage(ctx context.Context, message *session_grpc.Message) (*session_grpc.Message, error) {
	messageDomain := domain.Message{
		SessionID:   message.SessionID,
		SessionType: int(message.SessionType),
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		Content:     message.Content,
	}

	created, err := server.svc.CreateMessage(ctx, messageDomain)
	if err != nil {
		return &session_grpc.Message{}, err
	}

	return dto.ToMessage(created), nil
}

func (server *SessionServiceServer) HealthCheck(ctx context.Context, req *session_grpc.HealthCheckRequest) (*session_grpc.HealthCheckResponse, error) {
	return &session_grpc.HealthCheckResponse{}, nil
}
