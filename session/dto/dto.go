package dto

import (
	"time"

	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"github.com/yzletter/go-postery/session/model"
)

type DTO struct {
	Content     string `json:"content"`
	MessageFrom int64  `json:"message_from,string"`
	MessageTo   int64  `json:"message_to,string"`
	ID          int64  `json:"id,string"`
	SessionID   int64  `json:"session_id,string"`
	SessionType int    `json:"session_type"`
	CreatedAt   string `json:"created_at"` // 创建时间
}

// ToMessage model.Message 转 session_grpc.Message
func ToMessage(message *model.Message) *session_grpc.Message {
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

// ToSession model 转 session_grpc.Session
func ToSession(session *model.Session) *session_grpc.Session {
	return &session_grpc.Session{
		//ID:              session.ID,
		SessionID:       session.SessionID,
		TargetID:        session.TargetID,
		LastMessageID:   session.LastMessageID,
		LastMessage:     session.LastMessage,
		LastMessageTime: session.UpdatedAt.Format(time.RFC3339),
		UnreadCount:     int64(session.UnreadCount),
	}
}
