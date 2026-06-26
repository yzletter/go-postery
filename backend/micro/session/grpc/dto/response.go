package dto

import (
	"time"

	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"github.com/yzletter/go-postery/backend/micro/session/domain"
)

// ToMessage domain.Message 转 session_grpc.Message
func ToMessage(message domain.Message) *session_grpc.Message {
	return &session_grpc.Message{
		ID:          message.ID,
		SessionID:   message.SessionID,
		SessionType: int32(message.SessionType),
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		Content:     message.Content,
		CreatedAt:   message.CreatedAt.Format(time.RFC3339),
	}
}

// ToSession domain.Session 转 session_grpc.Session
func ToSession(session domain.Session) *session_grpc.Session {
	return &session_grpc.Session{
		SessionID:       session.SessionID,
		TargetID:        session.TargetID,
		LastMessageID:   session.LastMessageID,
		LastMessage:     session.LastMessage,
		LastMessageTime: session.UpdatedAt.Format(time.RFC3339),
		UnreadCount:     int64(session.UnreadCount),
	}
}
