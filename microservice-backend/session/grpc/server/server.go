package server

import (
	"context"

	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"github.com/yzletter/go-postery/microservice-backend/session/dto"
	model2 "github.com/yzletter/go-postery/microservice-backend/session/model"
	"github.com/yzletter/go-postery/microservice-backend/session/service"
	"github.com/yzletter/go-postery/microservice-backend/session/utils"
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
	updates := model2.UpdateUnread{
		Updates: model2.Updates{
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
	messageModel := &model2.Message{
		SessionID:   message.SessionID,
		SessionType: int(message.SessionType),
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		Content:     message.Content,
	}

	created, err := server.svc.CreateMessage(ctx, messageModel)
	if err != nil {
		return &session_grpc.Message{}, err
	}

	return dto.ToMessage(created), nil
}

func (server *SessionServiceServer) HealthCheck(ctx context.Context, req *session_grpc.HealthCheckRequest) (*session_grpc.HealthCheckResponse, error) {
	return &session_grpc.HealthCheckResponse{}, nil
}
