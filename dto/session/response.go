package session

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

type DTO struct {
	ID              int64  `json:"id,string"`
	TargetID        int64  `json:"target_id"`
	TargetName      string `json:"target_name"`
	TargetAvatar    string `json:"target_avatar"`
	LastMessageID   int64  `json:"last_message_id"`   // 最后一条消息的 ID
	LastMessage     string `json:"last_message"`      // 最后一条消息的摘要
	LastMessageTime string `json:"last_message_time"` // 最后一条消息的时间
	UnreadCount     int    `json:"unread_count"`      // 未读消息数
}

func ToDTO(session *model.Session, user *model.User) DTO {
	return DTO{
		ID:              session.SessionID,
		TargetID:        session.TargetID,
		TargetName:      user.Username,
		TargetAvatar:    user.Avatar,
		LastMessageID:   session.LastMessageID,
		LastMessage:     session.LastMessage,
		LastMessageTime: session.UpdatedAt.Format(time.RFC3339),
		UnreadCount:     session.UnreadCount,
	}
}
