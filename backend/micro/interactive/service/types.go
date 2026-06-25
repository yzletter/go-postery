package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
)

// InteractiveService 定义互动服务的业务能力
type InteractiveService interface {
	// GetPostInteractive 获取帖子互动信息
	//
	// Parameter:
	//	- id: 帖子 ID
	//
	// Return:
	//	- domain.PostInter: 帖子的互动信息
	//	- error: 可能返回的错误
	GetPostInteractive(ctx context.Context, id int64) (domain.PostInter, error)

	// GetUserInteractive 获取用户互动信息
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- domain.UserInter: 用户的互动信息
	//	- error: 可能返回的错误
	GetUserInteractive(ctx context.Context, id int64) (domain.UserInter, error)

	// Like 为用户创建帖子点赞关系
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- postID: 帖子 ID
	//
	// Return:
	//	- error: 可能返回的错误
	Like(ctx context.Context, userID int64, postID int64) error

	// Unlike 取消用户的帖子点赞关系
	//
	// Parameter:
	//	- postID: 帖子 ID
	//	- userID: 用户 ID
	//
	// Return:
	//	- error: 可能返回的错误
	Unlike(ctx context.Context, postID int64, userID int64) error

	// CheckLike 检查用户是否已点赞帖子
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- postID: 帖子 ID
	//
	// Return:
	//	- bool: 是否已点赞
	//	- error: 可能返回的错误
	CheckLike(ctx context.Context, userID int64, postID int64) (bool, error)

	// Follow 创建用户关注关系
	//
	// Parameter:
	//	- follower: 关注者 ID
	//	- followee: 被关注者 ID
	//
	// Return:
	//	- error: 可能返回的错误
	Follow(ctx context.Context, follower int64, followee int64) error

	// Unfollow 取消用户关注关系
	//
	// Parameter:
	//	- follower: 关注者 ID
	//	- followee: 被关注者 ID
	//
	// Return:
	//	- error: 可能返回的错误
	Unfollow(ctx context.Context, follower int64, followee int64) error

	// IfFollow 获取两个用户之间的关注关系
	//
	// Parameter:
	//	- follower: 关注者 ID
	//	- followee: 被关注者 ID
	//
	// Return:
	//	- int: 关注关系
	//	- error: 可能返回的错误
	IfFollow(ctx context.Context, follower int64, followee int64) (int, error)

	// Comment 创建帖子评论
	//
	// Parameter:
	//	- postID: 帖子 ID
	//	- parentID: 父评论 ID
	//	- replyID: 被回复评论 ID
	//	- userID: 用户 ID
	//	- content: 评论内容
	//
	// Return:
	//	- domain.Comment: 评论
	//	- error: 可能返回的错误
	Comment(ctx context.Context, postID int64, parentID int64, replyID int64, userID int64, content string) (domain.Comment, error)

	// DelComment 删除评论
	//
	// Parameter:
	//	- commentID: 评论 ID
	//	- userID: 用户 ID
	//
	// Return:
	//	- error: 可能返回的错误
	DelComment(ctx context.Context, commentID int64, userID int64) error

	// ListCommentByPage 分页获取帖子的主评论
	//
	// Parameter:
	//	- postID: 帖子 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 主评论总数
	//	- []domain.Comment: 当前页的主评论
	//	- error: 可能返回的错误
	ListCommentByPage(ctx context.Context, postID int64, pageNo int, pageSize int) (int64, []domain.Comment, error)

	// ListRepliesByPage 分页获取评论的回复
	//
	// Parameter:
	//	- commentID: 评论 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 回复总数
	//	- []domain.Comment: 当前页的回复
	//	- error: 可能返回的错误
	ListRepliesByPage(ctx context.Context, commentID int64, pageNo int, pageSize int) (int64, []domain.Comment, error)

	// CheckCommentDelAuth 检查用户是否有权限删除评论
	//
	// Parameter:
	//	- commentID: 评论 ID
	//	- userID: 用户 ID
	//
	// Return:
	//	- bool: 是否有删除权限
	//	- error: 可能返回的错误
	CheckCommentDelAuth(ctx context.Context, commentID int64, userID int64) (bool, error)

	// GetFollowers 分页获取用户的粉丝 ID
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 粉丝总数
	//	- []int64: 当前页的粉丝 ID
	//	- error: 可能返回的错误
	GetFollowers(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []int64, error)

	// GetFollowees 分页获取用户关注的用户 ID
	//
	// Parameter:
	//	- userID: 用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 关注的人总数
	//	- []int64: 当前页关注的用户 ID
	//	- error: 可能返回的错误
	GetFollowees(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []int64, error)

	// StartKafkaConsumer 启动互动消息消费者
	//
	// Parameter:
	//	- ctx: 上下文
	StartKafkaConsumer(ctx context.Context)
}
