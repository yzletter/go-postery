package cache

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/post/model"
)

type PostCache interface {
	// ChangeInteractiveCnt 修改帖子互动计数
	//
	// Parameter:
	//	- pid: 帖子 ID
	//	- field: 计数字段
	//	- delta: 计数变化
	//
	// Return:
	//	- bool: 是否修改成功
	//	- error: 可能返回的错误
	ChangeInteractiveCnt(ctx context.Context, pid int64, field model.PostCntField, delta int) (bool, error)

	// SetInteractiveKey 设置帖子互动计数字段
	//
	// Parameter:
	//	- pid: 帖子 ID
	//	- fields: 计数字段列表
	//	- vals: 计数值列表
	SetInteractiveKey(ctx context.Context, pid int64, fields []model.PostCntField, vals []int)

	// SetScore 设置帖子分数
	//
	// Parameter:
	//	- pid: 帖子 ID
	//
	// Return:
	//	- error: 可能返回的错误
	SetScore(ctx context.Context, pid int64) error

	// CheckPostLikeTime 查询帖子点赞时间分数
	//
	// Parameter:
	//	- pid: 帖子 ID
	//
	// Return:
	//	- float64: 点赞时间分数
	//	- error: 可能返回的错误
	CheckPostLikeTime(ctx context.Context, pid int64) (float64, error)

	// ChangeScore 修改帖子分数
	//
	// Parameter:
	//	- pid: 帖子 ID
	//	- delta: 分数变化
	//
	// Return:
	//	- error: 可能返回的错误
	ChangeScore(ctx context.Context, pid int64, delta int) error

	// Top 获取帖子分数排行
	//
	// Return:
	//	- []int64: 帖子 ID 列表
	//	- []float64: 帖子分数列表
	//	- error: 可能返回的错误
	Top(ctx context.Context) ([]int64, []float64, error)

	// DeleteScore 删除帖子分数
	//
	// Parameter:
	//	- id: 帖子 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DeleteScore(ctx context.Context, id int64) error

	// SetPost 设置帖子缓存
	//
	// Parameter:
	//	- id: 帖子 ID
	//	- post: 帖子
	//
	// Return:
	//	- error: 可能返回的错误
	SetPost(ctx context.Context, id int64, post *model.Post) error

	// DelPost 删除帖子缓存
	//
	// Parameter:
	//	- id: 帖子 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DelPost(ctx context.Context, id int64) error

	// GetPost 获取帖子缓存
	//
	// Parameter:
	//	- id: 帖子 ID
	//
	// Return:
	//	- *model.Post: 帖子
	//	- error: 可能返回的错误
	GetPost(ctx context.Context, id int64) (*model.Post, error)

	// SetAuthorHomePage 设置作者主页帖子缓存
	//
	// Parameter:
	//	- authorID: 作者 ID
	//	- posts: 帖子列表
	//
	// Return:
	//	- error: 可能返回的错误
	SetAuthorHomePage(ctx context.Context, authorID int64, posts []*model.Post) error

	// DelAuthorHomePage 删除作者主页帖子缓存
	//
	// Parameter:
	//	- authorID: 作者 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DelAuthorHomePage(ctx context.Context, authorID int64) error

	// GetAuthorHomePage 获取作者主页帖子缓存
	//
	// Parameter:
	//	- authorID: 作者 ID
	//
	// Return:
	//	- []*model.Post: 作者主页帖子列表
	//	- error: 可能返回的错误
	GetAuthorHomePage(ctx context.Context, authorID int64) ([]*model.Post, error)
}
