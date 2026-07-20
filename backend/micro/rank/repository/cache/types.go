package cache

import "context"

type RankCache interface {
	// DeleteScore 删除业务主体分数
	//
	// Parameter:
	//	- biz: 业务类型
	//	- bizID: 业务主体 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DeleteScore(ctx context.Context, biz int, bizID int64) error

	// UpdateScore 更新分数并放入排行榜
	//
	// Parameter:
	//	- biz: 业务类型
	//	- bizID: 业务主体 ID
	//	- score: 分数
	//
	// Return:
	//	- error: 可能返回的错误
	UpdateScore(ctx context.Context, biz int, bizID int64, score int64) error

	// TopK 获取业务排行榜
	//
	// Parameter:
	//	- biz: 业务类型
	//	- k: 榜单数量
	//
	// Return:
	//	- []int64: 业务主体 ID 列表
	//	- []int64: 分数列表
	//	- error: 可能返回的错误
	TopK(ctx context.Context, biz int, k int) ([]int64, []int64, error)
}
