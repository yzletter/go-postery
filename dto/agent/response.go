package agent

import (
	"github.com/cloudwego/eino/adk"
)

type DTO struct {
	SessionID int64  `json:"session_id,string"`
	Content   string `json:"content"`
}

func ToDTO(message adk.Message, ssid int64) DTO {
	return DTO{
		SessionID: ssid,
		Content:   message.Content,
	}
}
