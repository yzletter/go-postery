package dao

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/microservice-backend/lottery/model"
)

type OrderDAO interface {
	Create(ctx context.Context, order *model.Order) error
	Get(ctx context.Context, uid int64) (*model.Order, error)
	GetTempOrder(ctx context.Context, uid int64) (*model.Order, error)
	CreateTempOrder(ctx context.Context, order *model.Order) error                // 创建临时订单
	RecycleTempOrder(ctx context.Context, uid int64, orderID int64) (bool, error) // 回收临时订单，返回是否需要回补库存
	MarkRollbackDone(ctx context.Context, orderID int64) error
	MarkRollbackFailed(ctx context.Context, orderID int64, nextRollbackAt time.Time) error
	ListRollbackDueOrders(ctx context.Context, limit int) ([]*model.Order, error)
	PayTempOrder(ctx context.Context, orderID int64) error
	CancelTempOrder(ctx context.Context, orderID int64) error
	//GetTempOrder(ctx context.Context, uid int64) (*model.Order, error)
}

type GiftDAO interface {
	GetAll(ctx context.Context) ([]*model.Gift, error)
	GetByID(ctx context.Context, gid int64) (*model.Gift, error)
}
