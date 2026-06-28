package service

import "context"

type AgentService interface {
	// Chat 发起智能体对话
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- sessionID: 会话 ID
	//	- query: 用户问题
	//
	// Return:
	//	- string: 回答内容
	//	- []string: 引用内容列表
	//	- error: 可能返回的错误
	Chat(ctx context.Context, userID int64, sessionID int64, query string) (string, []string, error)

	// StartChunkDocConsumer 启动文档切块消费者
	//
	// Parameter:
	//	- ctx: 上下文
	StartChunkDocConsumer(ctx context.Context)

	// StartUpsertQdrantConsumer 启动 Qdrant 写入消费者
	//
	// Parameter:
	//	- ctx: 上下文
	StartUpsertQdrantConsumer(ctx context.Context)
}
