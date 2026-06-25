package repository

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/micro/post/domain"
	"github.com/yzletter/go-postery/backend/micro/post/model"
)

type PostRepository interface {
	// Create 创建帖子并绑定标签
	//
	// Parameter:
	//	- post: 帖子
	//	- events: 事件列表
	//
	// Return:
	//	- error: 可能返回的错误
	Create(ctx context.Context, post domain.Post, events []*event.OutboxEvent) error

	// Delete 删除帖子
	//
	// Parameter:
	//	- postID: 帖子 ID
	//	- authorID: 作者 ID
	//
	// Return:
	//	- error: 可能返回的错误
	Delete(ctx context.Context, postID int64, authorID int64) error

	// Update 更新帖子和标签
	//
	// Parameter:
	//	- post: 待更新帖子
	//
	// Return:
	//	- error: 可能返回的错误
	Update(ctx context.Context, post domain.Post) error

	// GetByID 根据 ID 获取帖子
	//
	// Parameter:
	//	- id: 帖子 ID
	//
	// Return:
	//	- domain.Post: 帖子
	//	- error: 可能返回的错误
	GetByID(ctx context.Context, id int64) (domain.Post, error)

	// GetPostByTime 根据时间获取帖子 ID
	//
	// Parameter:
	//	- timeAt: 查询起始时间
	//
	// Return:
	//	- []int64: 帖子 ID 列表
	//	- error: 可能返回的错误
	GetPostByTime(ctx context.Context, timeAt time.Time) ([]int64, error)

	// GetAuthorHomePage 获取作者主页帖子
	//
	// Parameter:
	//	- userID: 用户 ID
	//
	// Return:
	//	- []domain.Post: 作者主页帖子列表
	//	- error: 可能返回的错误
	GetAuthorHomePage(ctx context.Context, userID int64) ([]domain.Post, error)

	// GetByAuthor 按作者分页获取帖子
	//
	// Parameter:
	//	- id: 作者 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 帖子总数
	//	- []domain.Post: 当前页的帖子
	//	- error: 可能返回的错误
	GetByAuthor(ctx context.Context, id int64, pageNo, pageSize int) (int64, []domain.Post, error)

	// GetByPage 按页获取帖子
	//
	// Parameter:
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 帖子总数
	//	- []domain.Post: 当前页的帖子
	//	- error: 可能返回的错误
	GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []domain.Post, error)

	// GetByPageAndTag 按标签分页获取帖子
	//
	// Parameter:
	//	- tid: 标签 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 帖子总数
	//	- []domain.Post: 当前页的帖子
	//	- error: 可能返回的错误
	GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []domain.Post, error)
}

//type PostRepository interface {
//	Create(ctx context.Context, post *model.Post, events []*event.OutboxEvent) error
//	Delete(ctx context.Context, id int64) error
//	UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error
//	Update(ctx context.Context, id int64, updates map[string]any) error
//	GetByID(ctx context.Context, id int64) (*model.Post, error)
//	GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error)
//	GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error)
//	GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error)
//	ChangeScore(ctx context.Context, pid int64, delta int)
//	Top(ctx context.Context) ([]*model.Post, []float64, error)
//}

type TagRepository interface {
	// Create 创建标签
	//
	// Parameter:
	//	- tag: 标签
	//
	// Return:
	//	- error: 可能返回的错误
	Create(ctx context.Context, tag *model.Tag) error

	// GetBySlug 根据 slug 获取标签
	//
	// Parameter:
	//	- slug: 标签 slug
	//
	// Return:
	//	- *model.Tag: 标签
	//	- error: 可能返回的错误
	GetBySlug(ctx context.Context, slug string) (*model.Tag, error)

	// GetByName 根据名称获取标签
	//
	// Parameter:
	//	- name: 标签名称
	//
	// Return:
	//	- *model.Tag: 标签
	//	- error: 可能返回的错误
	GetByName(ctx context.Context, name string) (*model.Tag, error)

	// Bind 绑定帖子和标签
	//
	// Parameter:
	//	- postTag: 帖子标签关系
	//
	// Return:
	//	- error: 可能返回的错误
	Bind(ctx context.Context, postTag *model.PostTag) error

	// DeleteBind 删除帖子和标签绑定
	//
	// Parameter:
	//	- pid: 帖子 ID
	//	- tid: 标签 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DeleteBind(ctx context.Context, pid, tid int64) error

	// FindTagsByPostID 根据帖子 ID 查找标签
	//
	// Parameter:
	//	- pid: 帖子 ID
	//
	// Return:
	//	- []string: 标签名称列表
	//	- error: 可能返回的错误
	FindTagsByPostID(ctx context.Context, pid int64) ([]string, error)
}
