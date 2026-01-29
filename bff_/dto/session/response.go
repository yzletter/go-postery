package session

import (
	"time"

	model2 "github.com/yzletter/go-postery/auth/model"
	"github.com/yzletter/go-postery/session/model"
)

type DTO struct {
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

func ToDTO(session *model.Session, userProfile *model2.UserProfile) DTO {
	var res = DTO{
		//ID:              session.ID,
		SessionID:       session.SessionID,
		TargetID:        session.TargetID,
		TargetName:      userProfile.Nickname,
		TargetAvatar:    "",
		LastMessageID:   session.LastMessageID,
		LastMessage:     session.LastMessage,
		LastMessageTime: session.UpdatedAt.Format(time.RFC3339),
		UnreadCount:     session.UnreadCount,
	}

	if userProfile.Avatar != nil {
		res.TargetAvatar = *userProfile.Avatar
	}

	return res
}

type DTO struct {
	Content     string `json:"content"`
	MessageFrom int64  `json:"message_from,string"`
	MessageTo   int64  `json:"message_to,string"`
	ID          int64  `json:"id,string"`
	SessionID   int64  `json:"session_id,string"`
	SessionType int    `json:"session_type"`
	CreatedAt   string `json:"created_at"` // 创建时间
}

func ToDTO(message *model.Message) DTO {
	return DTO{
		Content:     message.Content,
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		ID:          message.ID,
		SessionID:   message.SessionID,
		SessionType: message.SessionType,
		CreatedAt:   message.CreatedAt.Format(time.RFC3339),
	}
}
