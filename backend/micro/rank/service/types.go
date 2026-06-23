package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/rank/domain"
)

type RankService interface {
	// RankUser 计算用户分数
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- error: 可能返回的错误
	RankUser(ctx context.Context, id int64) error

	// RankPost 计算文章分数
	//
	// Parameter:
	//	- id: 帖子 ID
	//
	// Return:
	//	- error: 可能返回的错误
	RankPost(ctx context.Context, id int64) error

	// RankTopKUser 计算用户榜单
	//
	// Return:
	//	- error: 可能返回的错误
	RankTopKUser(ctx context.Context) error

	// RankTopKPost 计算文章榜单
	//
	// Return:
	//	- error: 可能返回的错误
	RankTopKPost(ctx context.Context) error

	// TopKUser 返回用户榜单
	//
	// Return:
	//	- []domain.User: 用户榜单
	//	- error: 可能返回的错误
	TopKUser(ctx context.Context) ([]domain.User, error)

	// TopKPost 返回文章榜单
	//
	// Return:
	//	- []domain.Post: 文章榜单
	//	- error: 可能返回的错误
	TopKPost(ctx context.Context) ([]domain.Post, error)

	// StartKafkaConsumer 启动排行榜消息消费者
	//
	// Parameter:
	//	- ctx: 上下文
	StartKafkaConsumer(ctx context.Context)

	CronRankTopK()
}
