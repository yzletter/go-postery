package cache

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/lottery/model"
)

// GiftCache 奖品缓存接口
type GiftCache interface {
	// InitInventory 初始化库存缓存
	//
	// Parameter:
	//	- gifts: 奖品列表
	InitInventory(ctx context.Context, gifts []*model.Gift)

	// GetAllInventory 获取所有库存缓存
	//
	// Return:
	//	- []*model.Gift: 奖品库存列表
	//	- error: 可能返回的错误
	GetAllInventory(ctx context.Context) ([]*model.Gift, error)

	// Lottery 预扣库存并保存或复用抽奖预占记录
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

	// RollbackInventory 回滚库存缓存
	//
	// Parameter:
	//	- orderID: 订单 ID
	//	- gid: 奖品 ID
	//
	// Return:
	//	- error: 可能返回的错误
	RollbackInventory(ctx context.Context, orderID, gid int64) error
}
