package dao

import (
	"context"

	"github.com/yzletter/go-postery/backend/event"
	model2 "github.com/yzletter/go-postery/backend/event/outbox/model"
	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
	"github.com/yzletter/go-postery/backend/micro/interactive/model"
)

// InteractiveDAO 定义互动模块的数据访问能力
type InteractiveDAO interface {
	// GetPostInteractive 根据帖子 ID 获取帖子互动数据
	//
	// Parameter:
	//	- id: 帖子 ID
	//
	// Return:
	//	- model.PostInteractive: 帖子互动数据
	//	- error: 可能返回的错误
	GetPostInteractive(ctx context.Context, id int64) (model.PostInteractive, error)

	// GetUserInteractive 根据用户 ID 获取用户互动数据
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- model.UserInteractive: 用户互动数据
	//	- error: 可能返回的错误
	GetUserInteractive(ctx context.Context, id int64) (model.UserInteractive, error)

	// IncrReadCnt 增加帖子的阅读数
	//
	// Parameter:
	//	- id: 帖子 ID
	//	- cnt: 阅读数变化
	//
	// Return:
	//	- error: 可能返回的错误
	IncrReadCnt(ctx context.Context, consumer string, topic string, readEventPayloads ...*event.NewReadEventPayload) error

	// ChangeInteractiveCnt 修改互动数据
	//
	// Parameter:
	//	- biz: 业务区分
	//	- bizID: 业务主体 ID
	//	- delta: 互动数变化
	//
	// Return:
	//	- error: 可能返回的错误
	ChangeInteractiveCnt(ctx context.Context, biz domain.BizType, bizID int64, delta int64, processedEvent *model2.ProcessedEvent) error

	// CreateLike 创建点赞记录
	//
	// Parameter:
	//	- like: 点赞记录
	//
	// Return:
	//	- error: 可能返回的错误
	CreateLike(ctx context.Context, like *model.Like, events ...*model2.OutboxEvent) error

	// DelLike 删除点赞记录
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- pid: 帖子 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DelLike(ctx context.Context, uid, pid int64, events ...*model2.OutboxEvent) error

	// GetLike 查询用户是否已点赞帖子
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- pid: 帖子 ID
	//
	// Return:
	//	- bool: 是否已点赞
	//	- error: 可能返回的错误
	GetLike(ctx context.Context, uid, pid int64) (bool, error)

	// CreateFollow 创建关注关系
	//
	// Parameter:
	//	- follow: 关注关系
	//
	// Return:
	//	- error: 可能返回的错误
	CreateFollow(ctx context.Context, follow *model.Follow, events ...*model2.OutboxEvent) error

	// DelFollow 删除关注关系
	//
	// Parameter:
	//	- ferID: 关注者 ID
	//	- feeID: 被关注者 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DelFollow(ctx context.Context, ferID, feeID int64, events ...*model2.OutboxEvent) error

	// GetFollow 获取两个用户之间的关注关系
	//
	// Parameter:
	//	- follower: 关注者 ID
	//	- followee: 被关注者 ID
	//
	// Return:
	//	- domain.FollowType: 关注关系
	//	- error: 可能返回的错误
	GetFollow(ctx context.Context, follower, followee int64) (domain.FollowType, error)

	// GetFollowers 分页获取用户的粉丝 ID
	//
	// Parameter:
	//	- id: 用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 粉丝总数
	//	- []int64: 当前页的粉丝 ID
	//	- error: 可能返回的错误
	GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)

	// GetFollowees 分页获取用户关注的用户 ID
	//
	// Parameter:
	//	- id: 用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 关注的人总数
	//	- []int64: 当前页关注的用户 ID
	//	- error: 可能返回的错误
	GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)

	// CreateComment 创建评论记录
	//
	// Parameter:
	//	- comment: 评论记录
	//
	// Return:
	//	- error: 可能返回的错误
	CreateComment(ctx context.Context, comment *model.Comment, events ...*model2.OutboxEvent) (*model.Comment, error)

	// GetCommentByID 根据评论 ID 获取评论
	//
	// Parameter:
	//	- id: 评论 ID
	//
	// Return:
	//	- model.Comment: 评论
	//	- error: 可能返回的错误
	GetCommentByID(ctx context.Context, id int64) (model.Comment, error)

	// DelComment 删除评论并返回受影响的评论数量
	//
	// Parameter:
	//	- id: 评论 ID
	//
	// Return:
	//	- int: 受影响的评论数量
	//	- error: 可能返回的错误
	DelComment(ctx context.Context, id int64, buildEvents func(cnt int) []*model2.OutboxEvent) (int, error)

	// GetCommentByPostID 分页获取帖子的主评论
	//
	// Parameter:
	//	- id: 帖子 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 主评论总数
	//	- []model.Comment: 当前页的主评论
	//	- error: 可能返回的错误
	GetCommentByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []model.Comment, error)

	// GetCommentRepliesByParentID 分页获取评论的回复
	//
	// Parameter:
	//	- id: 评论 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 回复总数
	//	- []model.Comment: 当前页的回复
	//	- error: 可能返回的错误
	GetCommentRepliesByParentID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []model.Comment, error)
}
