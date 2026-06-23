package repository

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/rank/domain"
)

type RankRepository interface {
	// UpdateUserScore 更新用户分数并放入排行榜
	//
	// Parameter:
	//	- id: 用户 ID
	//	- score: 用户分数
	//
	// Return:
	//	- error: 可能返回的错误
	UpdateUserScore(ctx context.Context, id int64, score int64) error

	// UpdatePostScore 更新帖子分数并放入排行榜
	//
	// Parameter:
	//	- id: 帖子 ID
	//	- score: 帖子分数
	//
	// Return:
	//	- error: 可能返回的错误
	UpdatePostScore(ctx context.Context, id int64, score int64) error

	// GetTopKUser 返回用户榜单
	//
	// Return:
	//	- []domain.User: 用户榜单
	//	- error: 可能返回的错误
	GetTopKUser(ctx context.Context) ([]domain.User, error)

	// GetTopKPost 返回文章榜单
	//
	// Return:
	//	- []domain.Post: 文章榜单
	//	- error: 可能返回的错误
	GetTopKPost(ctx context.Context) ([]domain.Post, error)
}
