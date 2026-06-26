package repository

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/session/domain"
)

type SessionRepository interface {
	// Create 创建会话
	//
	// Parameter:
	//	- session: 会话列表
	//
	// Return:
	//	- error: 可能返回的错误
	Create(ctx context.Context, session ...domain.Session) error

	// Recover 恢复单边删除的会话
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- targetID: 目标用户 ID
	//
	// Return:
	//	- domain.Session: 会话
	//	- error: 可能返回的错误
	Recover(ctx context.Context, uid, targetID int64) (domain.Session, error)

	// ListByUid 根据用户 ID 获取会话列表
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- []domain.Session: 会话列表
	//	- error: 可能返回的错误
	ListByUid(ctx context.Context, uid int64) ([]domain.Session, error)

	// GetByUidAndTargetID 根据用户 ID 和目标用户 ID 获取会话
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- targetID: 目标用户 ID
	//
	// Return:
	//	- domain.Session: 会话
	//	- error: 可能返回的错误
	GetByUidAndTargetID(ctx context.Context, uid, targetID int64) (domain.Session, error)

	// GetByID 根据 ID 获取会话
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- sid: 会话 ID
	//
	// Return:
	//	- domain.Session: 会话
	//	- error: 可能返回的错误
	GetByID(ctx context.Context, uid, sid int64) (domain.Session, error)

	// Delete 删除会话
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- sid: 会话 ID
	//
	// Return:
	//	- error: 可能返回的错误
	Delete(ctx context.Context, uid, sid int64) error

	// UpdateUnread 更新未读数
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- sid: 会话 ID
	//	- updates: 未读数更新信息
	//
	// Return:
	//	- error: 可能返回的错误
	UpdateUnread(ctx context.Context, uid int64, sid int64, updates domain.UpdateUnread) error

	// ClearUnread 清除未读数
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- sid: 会话 ID
	//
	// Return:
	//	- error: 可能返回的错误
	ClearUnread(ctx context.Context, uid int64, sid int64) error
}

type MessageRepository interface {
	// Create 创建消息
	//
	// Parameter:
	//	- message: 消息
	//
	// Return:
	//	- error: 可能返回的错误
	Create(ctx context.Context, message domain.Message) error

	// GetByIDAndTargetID 根据用户 ID 和目标用户 ID 获取消息
	//
	// Parameter:
	//	- id: 用户 ID
	//	- targetID: 目标用户 ID
	//
	// Return:
	//	- []domain.Message: 消息列表
	//	- error: 可能返回的错误
	GetByIDAndTargetID(ctx context.Context, id, targetID int64) ([]domain.Message, error)

	// GetByPage 分页获取消息
	//
	// Parameter:
	//	- id: 用户 ID
	//	- targetID: 目标用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int: 消息总数
	//	- []domain.Message: 当前页的消息
	//	- error: 可能返回的错误
	GetByPage(ctx context.Context, id int64, targetID int64, pageNo, pageSize int) (int, []domain.Message, error)
}
