package dto

import (
	"time"

	"github.com/yzletter/go-postery/backend/micro/session/domain"
)

type MessageDTO struct {
	Content     string `json:"content"`
	MessageFrom int64  `json:"message_from,string"`
	MessageTo   int64  `json:"message_to,string"`
	ID          int64  `json:"id,string"`
	SessionID   int64  `json:"session_id,string"`
	SessionType int    `json:"session_type"`
	CreatedAt   string `json:"created_at"` // 创建时间
}

func ToMessageDTO(message *domain.Message) MessageDTO {
	return MessageDTO{
		Content:     message.Content,
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		ID:          message.ID,
		SessionID:   message.SessionID,
		SessionType: message.SessionType,
		CreatedAt:   message.CreatedAt.Format(time.RFC3339),
	}
}
