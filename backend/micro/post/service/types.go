package service

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/micro/post/domain"
)

type PostService interface {
	// Create 创建帖子
	//
	// Parameter:
	//	- post: 帖子
	//
	// Return:
	//	- domain.Post: 帖子
	//	- error: 可能返回的错误
	Create(ctx context.Context, post domain.Post) (domain.Post, error)

	// GetDetailByID 根据帖子 ID 查询帖子详情
	//
	// Parameter:
	//	- postID: 帖子 ID
	//	- addViewCnt: 是否增加浏览数
	//
	// Return:
	//	- domain.Post: 帖子详情
	//	- error: 可能返回的错误
	GetDetailByID(ctx context.Context, postID int64, addViewCnt bool) (domain.Post, error)

	// GetBriefByID 根据帖子 ID 查询帖子摘要
	//
	// Parameter:
	//	- postID: 帖子 ID
	//
	// Return:
	//	- domain.PostBrief: 帖子摘要
	//	- error: 可能返回的错误
	GetBriefByID(ctx context.Context, postID int64) (domain.PostBrief, error)

	// GetPostByTime 根据时间获取帖子 ID
	//
	// Parameter:
	//	- timeAt: 查询起始时间
	//
	// Return:
	//	- []int64: 帖子 ID 列表
	//	- error: 可能返回的错误
	GetPostByTime(ctx context.Context, timeAt time.Time) ([]int64, error)

	// Top 返回推荐帖子
	//
	// Return:
	//	- []domain.PostBrief: 帖子摘要列表
	//	- []float64: 帖子分数列表
	//	- error: 可能返回的错误
	Top(ctx context.Context) ([]domain.PostBrief, []float64, error)

	// Update 更新帖子和标签
	//
	// Parameter:
	//	- post: 帖子
	//
	// Return:
	//	- error: 可能返回的错误
	Update(ctx context.Context, post domain.Post) error

	// ListByPage 按页查询帖子
	//
	// Parameter:
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 帖子总数
	//	- []domain.Post: 当前页的帖子
	//	- error: 可能返回的错误
	ListByPage(ctx context.Context, pageNo int, pageSize int) (int64, []domain.Post, error)

	// ListByPageAndUid 按页和用户 ID 查询用户发表的帖子
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 帖子总数
	//	- []domain.Post: 当前页的帖子
	//	- error: 可能返回的错误
	ListByPageAndUid(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []domain.Post, error)

	// ListByPageAndTag 按页和 Tag 查询帖子
	//
	// Parameter:
	//	- tag: 标签
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 帖子总数
	//	- []domain.Post: 当前页的帖子
	//	- error: 可能返回的错误
	ListByPageAndTag(ctx context.Context, tag string, pageNo int, pageSize int) (int64, []domain.Post, error)

	// Belong 查询帖子是否属于用户
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- postID: 帖子 ID
	//
	// Return:
	//	- bool: 是否属于该用户
	//	- error: 可能返回的错误
	Belong(ctx context.Context, userID int64, postID int64) (bool, error)

	// Delete 删除帖子
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- postID: 帖子 ID
	//
	// Return:
	//	- error: 可能返回的错误
	Delete(ctx context.Context, userID int64, postID int64) error
}
