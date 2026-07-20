package dao

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/micro/post/model"
)

// PostDAO 定义帖子 DAO 接口
type PostDAO interface {
	// Create 创建帖子
	//
	// Parameter:
	//	- post: 帖子
	//	- tags: 标签列表
	//	- postTags: 帖子标签关系列表
	//	- events: 事件列表
	//
	// Return:
	//	- error: 可能返回的错误
	//		- ErrParamsInvalid: 参数错误
	//		- ErrServerInternal: 数据库内部错误
	//		- ErrUniqueKey: 唯一键冲突
	Create(ctx context.Context, post *model.Post, tags []*model.Tag, postTags []*model.PostTag, events []*event.OutboxEvent) error

	// Delete 删除帖子
	//
	// Parameter:
	//	- id: 帖子 ID
	//	- events: 事件列表
	//
	// Return:
	//	- error: 可能返回的错误
	Delete(ctx context.Context, id int64, events []*event.OutboxEvent) error

	// UpdateCount 更新帖子计数字段
	//
	// Parameter:
	//	- id: 帖子 ID
	//	- field: 计数字段
	//	- delta: 计数变化
	//
	// Return:
	//	- error: 可能返回的错误
	UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error

	// Update 更新帖子和标签
	//
	// Parameter:
	//	- post: 待更新帖子
	//	- tags: 待绑定标签
	//	- postTags: 帖子标签关系
	//	- events: 事件列表
	//
	// Return:
	//	- error: 可能返回的错误
	//		- ErrParamsInvalid: 参数错误
	//		- ErrServerInternal: 数据库内部错误
	//		- ErrRecordNotFound: 帖子不存在
	//		- ErrUniqueKey: 唯一键冲突
	Update(ctx context.Context, post *model.Post, tags []*model.Tag, postTags []*model.PostTag, events []*event.OutboxEvent) error

	// GetByID 根据 ID 获取帖子
	//
	// Parameter:
	//	- id: 帖子 ID
	//
	// Return:
	//	- *model.Post: 帖子
	//	- error: 可能返回的错误
	GetByID(ctx context.Context, id int64) (*model.Post, error)

	// GetPostByTime 根据时间获取帖子 ID
	//
	// Parameter:
	//	- timeAt: 查询起始时间
	//
	// Return:
	//	- []int64: 帖子 ID 列表
	//	- error: 可能返回的错误
	GetPostByTime(ctx context.Context, timeAt time.Time) ([]int64, error)

	// GetByUid 按用户 ID 分页获取帖子
	//
	// Parameter:
	//	- id: 用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 帖子总数
	//	- []*model.Post: 当前页的帖子
	//	- error: 可能返回的错误
	GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error)

	// GetByPage 按页获取帖子
	//
	// Parameter:
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 帖子总数
	//	- []*model.Post: 当前页的帖子
	//	- error: 可能返回的错误
	GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error)

	// GetByPageAndTag 按标签分页获取帖子
	//
	// Parameter:
	//	- tid: 标签 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 帖子总数
	//	- []*model.Post: 当前页的帖子
	//	- error: 可能返回的错误
	GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error)
}

type TagDAO interface {
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
