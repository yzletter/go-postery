package domain

import (
	"time"

	"github.com/yzletter/go-postery/backend/micro/session/model"
)

type Session struct {
	ID            int64
	SessionID     int64
	UserID        int64
	TargetID      int64
	TargetType    int
	LastMessageID int64
	LastMessage   string
	UnreadCount   int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Message struct {
	ID          int64
	SessionID   int64
	SessionType int
	MessageFrom int64
	MessageTo   int64
	Content     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpdateUnread 修改未读
type UpdateUnread struct {
	Updates Updates
	Delta   int
}

type Updates struct {
	LastMessageID   int64
	LastMessage     string
	LastMessageTime time.Time
}

// ToModelSession domain.Session 转 model.Session
func ToModelSession(session Session) *model.Session {
	return &model.Session{
		ID:            session.ID,
		SessionID:     session.SessionID,
		UserID:        session.UserID,
		TargetID:      session.TargetID,
		TargetType:    session.TargetType,
		LastMessageID: session.LastMessageID,
		LastMessage:   session.LastMessage,
		UnreadCount:   session.UnreadCount,
		CreatedAt:     session.CreatedAt,
		UpdatedAt:     session.UpdatedAt,
	}
}

// ToModelSessions []domain.Session 转 []*model.Session
func ToModelSessions(sessions ...Session) []*model.Session {
	res := make([]*model.Session, 0, len(sessions))
	for _, session := range sessions {
		res = append(res, ToModelSession(session))
	}
	return res
}

// ToDomainSession model.Session 转 domain.Session
func ToDomainSession(session *model.Session) Session {
	return Session{
		ID:            session.ID,
		SessionID:     session.SessionID,
		UserID:        session.UserID,
		TargetID:      session.TargetID,
		TargetType:    session.TargetType,
		LastMessageID: session.LastMessageID,
		LastMessage:   session.LastMessage,
		UnreadCount:   session.UnreadCount,
		CreatedAt:     session.CreatedAt,
		UpdatedAt:     session.UpdatedAt,
	}
}

// ToDomainSessions []*model.Session 转 []domain.Session
func ToDomainSessions(sessions []*model.Session) []Session {
	res := make([]Session, 0, len(sessions))
	for _, session := range sessions {
		res = append(res, ToDomainSession(session))
	}
	return res
}

// ToModelMessage domain.Message 转 model.Message
func ToModelMessage(message Message) model.Message {
	return model.Message{
		ID:          message.ID,
		SessionID:   message.SessionID,
		SessionType: message.SessionType,
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		Content:     message.Content,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}
}

// ToDomainMessage model.Message 转 domain.Message
func ToDomainMessage(message *model.Message) Message {
	return Message{
		ID:          message.ID,
		SessionID:   message.SessionID,
		SessionType: message.SessionType,
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		Content:     message.Content,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}
}

// ToDomainMessages []*model.Message 转 []domain.Message
func ToDomainMessages(messages []*model.Message) []Message {
	res := make([]Message, 0, len(messages))
	for _, message := range messages {
		res = append(res, ToDomainMessage(message))
	}
	return res
}

// ToModelUpdateUnread domain.UpdateUnread 转 model.UpdateUnread
func ToModelUpdateUnread(updates UpdateUnread) model.UpdateUnread {
	return model.UpdateUnread{
		Updates: model.Updates{
			LastMessageID:   updates.Updates.LastMessageID,
			LastMessage:     updates.Updates.LastMessage,
			LastMessageTime: updates.Updates.LastMessageTime,
		},
		Delta: updates.Delta,
	}
}
