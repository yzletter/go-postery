package session

import (
	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
)

type SessionDTO struct {
	//ID              int64  `json:"id,string"`
	SessionID       int64  `json:"session_id,string"`
	TargetID        int64  `json:"target_id,string"`
	TargetName      string `json:"target_name"`
	TargetAvatar    string `json:"target_avatar"`
	LastMessageID   int64  `json:"last_message_id,string"` // 最后一条消息的 ID
	LastMessage     string `json:"last_message"`           // 最后一条消息的摘要
	LastMessageTime string `json:"last_message_time"`      // 最后一条消息的时间
	UnreadCount     int    `json:"unread_count"`           // 未读消息数
}

func ToSessionDTO(session *session_grpc.Session, user *user_grpc.Profile) SessionDTO {
	var res = SessionDTO{
		//ID:              session.ID,
		SessionID:       session.SessionID,
		TargetID:        session.TargetID,
		TargetName:      user.Nickname,
		TargetAvatar:    user.Avatar,
		LastMessageID:   session.LastMessageID,
		LastMessage:     session.LastMessage,
		LastMessageTime: session.LastMessageTime,
		UnreadCount:     int(session.UnreadCount),
	}

	return res
}

type MessageDTO struct {
	Content     string `json:"content"`
	MessageFrom int64  `json:"message_from,string"`
	MessageTo   int64  `json:"message_to,string"`
	ID          int64  `json:"id,string"`
	SessionID   int64  `json:"session_id,string"`
	SessionType int    `json:"session_type"`
	CreatedAt   string `json:"created_at"` // 创建时间
}

func ToMessageDTO(message *session_grpc.Message) MessageDTO {
	return MessageDTO{
		Content:     message.Content,
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		ID:          message.ID,
		SessionID:   message.SessionID,
		SessionType: int(message.SessionType),
		CreatedAt:   message.CreatedAt,
	}
}
