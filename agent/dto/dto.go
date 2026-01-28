package dto

import (
	"github.com/cloudwego/eino/adk"
	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
)

// ToChatResponse adk.Message 转 agent_grpc.ChatResponse
func ToChatResponse(message adk.Message, ssid int64, knowledge []string) *agent_grpc.ChatResponse {
	response := &agent_grpc.ChatResponse{
		SessionID: ssid,
		Content:   "对不起，这个问题我还在学习中……",
		Documents: []string{},
	}
	
	if message != nil {
		response.Content = message.Content
	}
	if knowledge != nil {
		response.Documents = knowledge
	}
	return response
}
