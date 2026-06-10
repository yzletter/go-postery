package dao

import (
	"context"
	"time"

	model2 "github.com/yzletter/go-postery/micro-backend/lottery/model"
)

type OrderDAO interface {
	Create(ctx context.Context, order *model2.Order) error
	Get(ctx context.Context, uid int64) (*model2.Order, error)
	GetTempOrder(ctx context.Context, uid int64) (*model2.Order, error)
	CreateTempOrder(ctx context.Context, order *model2.Order) error               // 创建临时订单
	RecycleTempOrder(ctx context.Context, uid int64, orderID int64) (bool, error) // 回收临时订单，返回是否需要回补库存
	MarkRollbackDone(ctx context.Context, orderID int64) error
	MarkRollbackFailed(ctx context.Context, orderID int64, nextRollbackAt time.Time) error
	ListRollbackDueOrders(ctx context.Context, limit int) ([]*model2.Order, error)
	PayTempOrder(ctx context.Context, orderID int64) error
	CancelTempOrder(ctx context.Context, orderID int64) error
	//GetTempOrder(ctx context.Context, uid int64) (*model.Order, error)
}

type GiftDAO interface {
	GetAll(ctx context.Context) ([]*model2.Gift, error)
	GetByID(ctx context.Context, gid int64) (*model2.Gift, error)
}
