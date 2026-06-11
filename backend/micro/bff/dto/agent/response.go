package agent

import (
	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
)

type DTO struct {
	SessionID int64    `json:"session_id,string"`
	Content   string   `json:"content"`
	Documents []string `json:"documents"`
}

func ToDTO(message *agent_grpc.ChatResponse) DTO {
	return DTO{
		SessionID: message.SessionID,
		Content:   message.Content,
		Documents: message.Documents,
	}
}
