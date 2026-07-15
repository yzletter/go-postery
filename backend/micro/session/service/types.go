package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/session/domain"
)

type SessionService interface {
	// NewConnection 为用户连接启动消息队列消费，直到 ctx 被取消。
	NewConnection(ctx context.Context, userID int64) error

	// Chat 处理一条来自已认证用户的聊天消息。
	Chat(ctx context.Context, userID int64, message domain.Message) error

	// ListByUID 根据用户 ID 获取会话列表
	//
	// Parameter:
	//	- userID: 用户 ID
	//
	// Return:
	//	- []domain.Session: 会话列表
	//	- error: 可能返回的错误
	ListByUID(ctx context.Context, userID int64) ([]domain.Session, error)

	// GetSession 获取用户和目标用户之间的会话
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- targetID: 目标用户 ID
	//
	// Return:
	//	- domain.Session: 会话
	//	- error: 可能返回的错误
	GetSession(ctx context.Context, userID int64, targetID int64) (domain.Session, error)

	// GetHistoryMessagesByPage 分页获取历史消息
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- targetID: 目标用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 消息总数
	//	- []domain.Message: 当前页的消息
	//	- error: 可能返回的错误
	GetHistoryMessagesByPage(ctx context.Context, userID int64, targetID int64, pageNo int, pageSize int) (int64, []domain.Message, error)

	// Delete 删除会话
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- sessionID: 会话 ID
	//
	// Return:
	//	- error: 可能返回的错误
	Delete(ctx context.Context, userID int64, sessionID int64) error

	// UpdateUnread 更新未读数
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- sessionID: 会话 ID
	//	- updates: 未读数更新信息
	//
	// Return:
	//	- error: 可能返回的错误
	UpdateUnread(ctx context.Context, userID int64, sessionID int64, updates domain.UpdateUnread) error

	// ClearUnread 清除未读数
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- sessionID: 会话 ID
	//
	// Return:
	//	- error: 可能返回的错误
	ClearUnread(ctx context.Context, userID int64, sessionID int64) error

	// CreateMessage 创建消息
	//
	// Parameter:
	//	- message: 消息
	//
	// Return:
	//	- domain.Message: 消息
	//	- error: 可能返回的错误
	CreateMessage(ctx context.Context, message domain.Message) (domain.Message, error)

	// StartSessionRegisterConsumer 启动会话注册消费者
	//
	// Parameter:
	//	- ctx: 上下文
	StartSessionRegisterConsumer(ctx context.Context)
}
