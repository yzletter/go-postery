package cache

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
)

// InteractiveCache 定义互动模块的缓存访问能力
type InteractiveCache interface {
	// GetPostInteractive 从缓存获取帖子互动数据
	//
	// Parameter:
	//	- id: 帖子 ID
	//
	// Return:
	//	- domain.PostInter: 帖子互动数据
	//	- error: 可能返回的错误
	GetPostInteractive(ctx context.Context, id int64) (domain.PostInter, error)

	// SetPostInteractive 写入帖子互动数据缓存
	//
	// Parameter:
	//	- id: 帖子 ID
	//	- inter: 帖子互动数据
	//
	// Return:
	//	- error: 可能返回的错误
	SetPostInteractive(ctx context.Context, id int64, inter domain.PostInter) error

	// DelPostInteractive 删除帖子互动数据缓存
	//
	// Parameter:
	//	- id: 帖子 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DelPostInteractive(ctx context.Context, id int64) error

	// GetUserInteractive 从缓存获取用户互动数据
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- domain.UserInter: 用户互动数据
	//	- error: 可能返回的错误
	GetUserInteractive(ctx context.Context, id int64) (domain.UserInter, error)

	// SetUserInteractive 写入用户互动数据缓存
	//
	// Parameter:
	//	- id: 用户 ID
	//	- inter: 用户互动数据
	//
	// Return:
	//	- error: 可能返回的错误
	SetUserInteractive(ctx context.Context, id int64, inter domain.UserInter) error

	// DelUserInteractive 删除用户互动数据缓存
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DelUserInteractive(ctx context.Context, id int64) error

	// GetLike 从缓存查询用户是否已点赞帖子
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- pid: 帖子 ID
	//
	// Return:
	//	- bool: 是否已点赞
	//	- error: 可能返回的错误
	GetLike(ctx context.Context, uid, pid int64) (bool, error)

	// SetLike 写入用户点赞关系缓存
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- pid: 帖子 ID
	//
	// Return:
	//	- error: 可能返回的错误
	SetLike(ctx context.Context, uid, pid int64) error

	// DelLike 删除用户点赞关系缓存
	//
	// Parameter:
	//	- uid: 用户 ID
	//	- pid: 帖子 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DelLike(ctx context.Context, uid, pid int64) error

	// GetFollow 从缓存获取两个用户之间的关注关系
	//
	// Parameter:
	//	- follower: 关注者 ID
	//	- followee: 被关注者 ID
	//
	// Return:
	//	- domain.FollowType: 关注关系
	//	- error: 可能返回的错误
	GetFollow(ctx context.Context, follower, followee int64) (domain.FollowType, error)

	// SetFollow 写入两个用户之间的关注关系缓存
	//
	// Parameter:
	//	- follower: 关注者 ID
	//	- followee: 被关注者 ID
	//
	// Return:
	//	- error: 可能返回的错误
	SetFollow(ctx context.Context, follower, followee int64) error

	// DelFollow 删除两个用户之间的关注关系缓存
	//
	// Parameter:
	//	- follower: 关注者 ID
	//	- followee: 被关注者 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DelFollow(ctx context.Context, follower, followee int64) error

	SetConsume(ctx context.Context, consumer string, id int64) error

	GetConsume(ctx context.Context, consumer string, id int64) (bool, error)
}
