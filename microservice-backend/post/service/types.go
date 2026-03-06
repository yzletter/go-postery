package service

import (
	"context"

	model2 "github.com/yzletter/go-postery/microservice-backend/post/model"
)

type PostWithTags struct {
	Post *model2.Post
	Tags []string
}

type PostService interface {
	Create(ctx context.Context, userID int64, title string, content string, contentType int, tags []string) (*PostWithTags, error) // 创建帖子
	GetDetailByID(ctx context.Context, postID int64, addViewCnt bool) (*PostWithTags, error)                                       // 根据帖子 ID 查询帖子详情
	GetBriefByID(ctx context.Context, postID int64) (*model2.Post, error)                                                          // 根据帖子 ID 查询帖子摘要
	Top(ctx context.Context) ([]*model2.Post, []float64, error)                                                                    // 返回推荐帖子
	Update(ctx context.Context, userID int64, postID int64, title string, content string, tags []string) error                     // 更新帖子
	ListByPage(ctx context.Context, pageNo int, pageSize int) (int64, []*PostWithTags, error)                                      // 按页查询帖子
	ListByPageAndUid(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []*model2.Post, error)                   // 按页 + 用户 ID 查询用户发表的帖子
	ListByPageAndTag(ctx context.Context, tag string, pageNo int, pageSize int) (int64, []*PostWithTags, error)                    // 按页 + Tag 查询该 Tag 的帖子
	Belong(ctx context.Context, userID int64, postID int64) (bool, error)                                                          // 查询帖子是否属于用户
	Delete(ctx context.Context, userID int64, postID int64) error                                                                  // 删除帖子
	Like(ctx context.Context, userID int64, postID int64) error                                                                    // 点赞帖子
	Unlike(ctx context.Context, userID int64, postID int64) error                                                                  // 取消点赞帖子
	IfLike(ctx context.Context, userID int64, postID int64) (bool, error)                                                          // 查询用户是否点过赞
	CreateComment(ctx context.Context, postID int64, parentID int64, replyID int64, userID int64, content string) (*model2.Comment, error)
	DeleteComment(ctx context.Context, commentID int64, userID int64) error
	ListCommentByPage(ctx context.Context, postID int64, pageNo int, pageSize int) (int64, []*model2.Comment, error)    // 根据 PostID 按页获取文章主评论
	ListRepliesByPage(ctx context.Context, commentID int64, pageNo int, pageSize int) (int64, []*model2.Comment, error) // 根据 CommentID 按页获取评论的回复
	CheckCommentDeleteAuth(ctx context.Context, commentID int64, userID int64) (bool, error)                            // 用户是否有评论删除权限
}
