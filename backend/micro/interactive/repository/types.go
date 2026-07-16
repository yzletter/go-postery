package repository

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
)

type InteractiveRepository interface {
	// GetPostInteractive 获取帖子互动信息
	//
	// Parameter
	// 	- id: 帖子 ID
	//
	// Return
	// 	- domain.PostInter: 帖子的互动信息
	// 	- error: 可能返回的错误
	// 	 	- ErrServerInternal: 系统内部错误
	// 	- 没有互动记录时返回零值
	GetPostInteractive(ctx context.Context, id int64) (domain.PostInter, error)

	// GetUserInteractive 获取用户互动信息
	//
	// Parameter
	// 	- id: 用户 ID
	//
	// Return
	// 	- domain.UserInter: 用户的互动信息
	// 	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	//	- 没有互动记录时返回零值
	GetUserInteractive(ctx context.Context, id int64) (domain.UserInter, error)

	// IncrReadCnt 增加帖子阅读数
	//
	// Parameter
	//	- id: 帖子 ID
	//	- cnt: 阅读数变化
	// Return
	// 	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	//		- ErrRecordNotFound: 帖子找不到
	IncrReadCnt(ctx context.Context, consumer string, topic string, readEventPayloads ...*event.NewReadEventPayload) error

	// ChangeInteractiveCntWithOutbox 更改互动信息并事务写消费表
	//
	// Parameter
	//	- biz: 业务区分
	//	- bizID: 业务主体 ID
	//	- timeAt: 业务发生时间
	//	- delta: 阅读数变化
	// Return
	// 	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	//		- ErrRecordNotFound: 业务主题找不到
	ChangeInteractiveCntWithOutbox(ctx context.Context, biz domain.BizType, bizID int64, timeAt time.Time, delta int64,
		processedEvent *event.ProcessedEvent) error

	// Like 点赞
	//
	// Parameter
	//	- like: 点赞
	//
	// Return
	// 	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	//		- ErrUniqueKey: 重复点赞
	Like(ctx context.Context, like domain.Like, events ...*event.OutboxEvent) error

	// UnLike 取消点赞
	//
	// Parameter
	//	- uid: 用户 ID
	// 	- pid: 帖子 ID
	//
	// Return
	// 	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	//		- ErrRecordNotFound: 重复取消点赞
	UnLike(ctx context.Context, uid, pid int64, events ...*event.OutboxEvent) error

	// HasLiked 用户是否已点赞
	//
	// Parameter
	//	- uid: 用户 ID
	// 	- pid: 帖子 ID
	//
	// Return
	// 	- bool: 是否点过赞
	//		- true: 已点赞
	//		- false: 未点过赞
	// 	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	HasLiked(ctx context.Context, uid, pid int64) (bool, error)

	// CreateFollow 新建关注关系
	//
	// Parameter:
	//	- follow: 关注关系
	//
	// Return:
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	CreateFollow(ctx context.Context, follow domain.Follow, events ...*event.OutboxEvent) error

	// DelFollow 删除关注关系
	//
	// Parameter:
	//	- follower: 关注者 ID
	//	- followee: 被关注者 ID
	//
	// Return:
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	//		- ErrRecordNotFound: 关注关系找不到或已删除
	DelFollow(ctx context.Context, follower, followee int64, events ...*event.OutboxEvent) error

	// GetFollow 获取关注关系
	//
	// Parameter:
	//	- follower: 关注者 ID
	//	- followee: 被关注者 ID
	//
	// Return:
	//	- domain.FollowType: 可能返回的关注关系
	//		- FollowNone: 没有关注关系
	//		- FollowIFollow: 单方面关注了被关注者
	//		- FollowFollowMe: 对方关注了我
	//		- FollowMutual: 互相关注
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	GetFollow(ctx context.Context, follower, followee int64) (domain.FollowType, error)

	// GetFollowers 按页获取用户关注的人
	//
	// Parameter:
	//	- id: 用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 关注的人总数
	//	- []int64: 当前页的所有关注的人的 ID
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)

	// GetFollowees 按页获取用户粉丝
	//
	// Parameter:
	//	- id: 用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 粉丝总数
	//	- []int64: 当前页的所有粉丝的 ID
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)

	// CreateComment 创建评论
	//
	// Parameter:
	//	- comment: domain.Comment 评论
	//
	// Return:
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	//		- ErrUniqueKey: 评论的雪花 ID 冲突
	CreateComment(ctx context.Context, comment domain.Comment, events ...*event.OutboxEvent) (domain.Comment, error)

	// GetCommentByID 根据 ID 获取评论
	//
	// Parameter:
	//	- id: 评论 ID
	//
	// Return:
	//	- *domain.Comment: 评论
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	//		- ErrRecordNotFound: 评论找不到
	GetCommentByID(ctx context.Context, id int64) (domain.Comment, error)

	// DelComment 删除评论
	//
	// Parameter:
	//	- id: 评论 ID
	//
	// Return:
	//	- int: 删除评论的个数
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	DelComment(ctx context.Context, id int64, buildEventsFunc func(cnt int) []*event.OutboxEvent) (int, error)

	// GetCommentByPostID 根据帖子 ID 按页获取主评论
	//
	// Parameter:
	//	- id: 帖子 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 主评论总数
	//	- []*domain.Comment: 该帖子当前页的所有主评论
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	GetCommentByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []domain.Comment, error)

	// GetCommentRepliesByParentID 根据主评论 ID 按页获取子评论
	//
	// Parameter:
	//	- id: 帖子 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 子评论总数
	//	- []*domain.Comment: 该主评论当前页的所有子评论
	//	- error: 可能返回的错误
	//		- ErrServerInternal: 系统内部错误
	GetCommentRepliesByParentID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []domain.Comment, error)
}
