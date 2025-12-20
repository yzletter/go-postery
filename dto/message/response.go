package message

import (
	"time"

	"github.com/yzletter/go-postery/model"
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
