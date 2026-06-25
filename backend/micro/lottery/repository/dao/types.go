package dao

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

type OrderDAO interface {
	// Create 创建订单
	//
	// Parameter:
	//	- order: 订单
	//
	// Return:
	//	- error: 可能返回的错误
	Create(ctx context.Context, order *model.Order) error

	// Get 获取订单
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- *model.Order: 订单
	//	- error: 可能返回的错误
	Get(ctx context.Context, uid int64) (*model.Order, error)

	// GetTempOrder 获取临时订单
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- *model.Order: 临时订单
	//	- error: 可能返回的错误
	GetTempOrder(ctx context.Context, uid int64) (*model.Order, error)

	// CreateTempOrder 创建临时订单
	//
	// Parameter:
	//	- order: 订单
	//
	// Return:
	//	- error: 可能返回的错误
	CreateTempOrder(ctx context.Context, order *model.Order) error

	// RecycleTempOrder 回收临时订单
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- orderID: 订单 ID
	//
	// Return:
	//	- bool: 是否需要回补库存
	//	- error: 可能返回的错误
	RecycleTempOrder(ctx context.Context, uid int64, orderID int64) (bool, error)

	// MarkRollbackDone 标记库存回滚完成
	//
	// Parameter:
	//	- orderID: 订单 ID
	//
	// Return:
	//	- error: 可能返回的错误
	MarkRollbackDone(ctx context.Context, orderID int64) error

	// MarkRollbackFailed 标记库存回滚失败
	//
	// Parameter:
	//	- orderID: 订单 ID
	//	- nextRollbackAt: 下次回滚时间
	//
	// Return:
	//	- error: 可能返回的错误
	MarkRollbackFailed(ctx context.Context, orderID int64, nextRollbackAt time.Time) error

	// ListRollbackDueOrders 获取到期回滚订单
	//
	// Parameter:
	//	- limit: 查询数量
	//
	// Return:
	//	- []*model.Order: 订单列表
	//	- error: 可能返回的错误
	ListRollbackDueOrders(ctx context.Context, limit int) ([]*model.Order, error)

	// PayTempOrder 支付订单
	//
	// Parameter:
	//	- orderID: 订单 ID
	//
	// Return:
	//	- error: 可能返回的错误
	PayTempOrder(ctx context.Context, orderID int64) error

	// CancelTempOrder 取消临时订单
	//
	// Parameter:
	//	- orderID: 订单 ID
	//
	// Return:
	//	- error: 可能返回的错误
	CancelTempOrder(ctx context.Context, orderID int64) error
	//GetTempOrder(ctx context.Context, uid int64) (*model.Order, error)
}

type GiftDAO interface {
	// GetAll 获取所有奖品
	//
	// Return:
	//	- []*model.Gift: 奖品列表
	//	- error: 可能返回的错误
	GetAll(ctx context.Context) ([]*model.Gift, error)

	// GetByID 根据 ID 获取奖品
	//
	// Parameter:
	//	- gid: 奖品 ID
	//
	// Return:
	//	- *model.Gift: 奖品
	//	- error: 可能返回的错误
	GetByID(ctx context.Context, gid int64) (*model.Gift, error)
}
