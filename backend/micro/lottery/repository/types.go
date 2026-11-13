package repository

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

// OrderRepository 订单 Repository 接口
type OrderRepository interface {
	// CreateOrder 创建订单，已存在时返回已有订单
	//
	// Parameter:
	//	- order: 订单
	//
	// Return:
	//	- *model.Order: 新建或已有的订单
	//	- error: 可能返回的错误
	CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error)

	// Pay 支付订单
	//
	// Parameter:
	//	- orderID: 订单 ID
	//
	// Return:
	//	- error: 可能返回的错误
	Pay(ctx context.Context, orderID int64) error

	// CancelTempOrder 取消临时订单
	//
	// Parameter:
	//	- orderID: 订单 ID
	//
	// Return:
	//	- error: 可能返回的错误
	CancelTempOrder(ctx context.Context, orderID int64) error

	// HandleTempOrder 回收临时订单
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- orderID: 订单 ID
	//
	// Return:
	//	- bool: 是否需要回补库存
	//	- error: 可能返回的错误
	HandleTempOrder(ctx context.Context, uid int64, orderID int64) (bool, error)

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

	// GetTempOrder 获取临时订单
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- *model.Order: 临时订单
	//	- error: 可能返回的错误
	GetTempOrder(ctx context.Context, uid int64) (*model.Order, error)

	// GetOrder 获取订单
	//
	// Parameter:
	//	- uid: 用户 ID
	//
	// Return:
	//	- *model.Order: 订单
	//	- error: 可能返回的错误
	GetOrder(ctx context.Context, uid int64) (*model.Order, error)
}

// GiftRepository 奖品 Repository 接口
type GiftRepository interface {
	// GetAllGifts 获取所有奖品
	//
	// Return:
	//	- []*model.Gift: 奖品列表
	//	- error: 可能返回的错误
	GetAllGifts(ctx context.Context) ([]*model.Gift, error)

	// GetCacheInventory 获取缓存库存
	//
	// Return:
	//	- []*model.Gift: 奖品库存列表
	//	- error: 可能返回的错误
	GetCacheInventory(ctx context.Context) ([]*model.Gift, error)

	// GetByID 根据 ID 获取奖品
	//
	// Parameter:
	//	- gid: 奖品 ID
	//
	// Return:
	//	- *model.Gift: 奖品
	//	- error: 可能返回的错误
	GetByID(ctx context.Context, gid int64) (*model.Gift, error)

	// Lottery 预扣缓存库存并保存或复用抽奖预占记录
	//
	// Parameter:
	//	- requestID: 请求 ID
	//	- userID: 用户 ID
	//	- giftID: 奖品 ID
	//
	// Return:
	//	- int64: 实际预占请求 ID
	//	- int64: 实际预占奖品 ID
	//	- error: 可能返回的错误
	Lottery(ctx context.Context, requestID int64, userID int64, giftID int64) (int64, int64, error)

	// Reuse 复用已有抽奖预占记录
	//
	// Parameter:
	//	- requestID: 请求 ID
	//	- userID: 用户 ID
	//
	// Return:
	//	- int64: 实际预占请求 ID
	//	- int64: 实际预占奖品 ID
	//	- error: 可能返回的错误
	Reuse(ctx context.Context, requestID int64, userID int64) (int64, int64, error)

	// ScanLotteryRecords 分页读取抽奖预占记录
	//
	// Parameter:
	//	- cursor: 当前扫描游标，首次扫描传 0
	//	- count: 每次扫描的数量提示
	//
	// Return:
	//	- []*model.LotteryRecord: 抽奖预占记录列表
	//	- uint64: 下次扫描游标，0 表示本轮扫描结束
	//	- error: 可能返回的错误
	ScanLotteryRecords(ctx context.Context, cursor uint64, count int64) ([]*model.LotteryRecord, uint64, error)

	// RollbackCacheInventory 回滚缓存库存
	//
	// Parameter:
	//	- orderID: 订单 ID
	//	- gid: 奖品 ID
	//
	// Return:
	//	- error: 可能返回的错误
	RollbackCacheInventory(ctx context.Context, orderID, gid int64) error

	// InitCacheInventory 初始化缓存库存
	//
	// Parameter:
	//	- ctx: 上下文
	InitCacheInventory(ctx context.Context)
}
