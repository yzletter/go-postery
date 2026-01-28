package service

import (
	"context"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
)

type PostService interface {
	Create(context.Context, *post_grpc.CreatePostRequest) (*post_grpc.PostDetail, error)                          // 创建帖子
	GetDetailByID(context.Context, *post_grpc.GetDetailByIDRequest) (*post_grpc.PostDetail, error)                // 根据帖子 ID 查询帖子详情
	GetBriefByID(context.Context, *post_grpc.GetBriefByIDRequest) (*post_grpc.PostBrief, error)                   // 根据帖子 ID 查询帖子摘要
	Top(context.Context, *post_grpc.PostEmptyRequest) (*post_grpc.TopResponse, error)                             // 返回推荐帖子
	Update(context.Context, *post_grpc.UpdateRequest) (*post_grpc.PostEmptyResponse, error)                       // 更新帖子
	ListByPage(context.Context, *post_grpc.ListByPageRequest) (*post_grpc.PostDetailsResponse, error)             // 按页查询帖子
	ListByPageAndUid(context.Context, *post_grpc.ListByPageAndUidRequest) (*post_grpc.PostBriefsResponse, error)  // 按页 + 用户 ID 查询用户发表的帖子
	ListByPageAndTag(context.Context, *post_grpc.ListByPageAndTagRequest) (*post_grpc.PostDetailsResponse, error) // 按页 + Tag 查询该 Tag 的帖子
	Belong(context.Context, *post_grpc.PostCommonRequest) (*post_grpc.BelongResponse, error)                      // 查询帖子是否属于用户
	Delete(context.Context, *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error)                   // 删除帖子
	Like(context.Context, *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error)                     // 点赞帖子
	Unlike(context.Context, *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error)                   // 取消点赞帖子
	IfLike(context.Context, *post_grpc.PostCommonRequest) (*post_grpc.IfLikeResponse, error)                      // 查询用户是否点过赞
	CreateComment(context.Context, *post_grpc.CreateCommentRequest) (*post_grpc.Comment, error)
	DeleteComment(context.Context, *post_grpc.DeleteCommentRequest) (*post_grpc.PostEmptyResponse, error)
	ListCommentByPage(context.Context, *post_grpc.ListCommentByPageRequest) (*post_grpc.CommentsResponse, error) // 根据 PostID 按页获取文章主评论
	ListRepliesByPage(context.Context, *post_grpc.ListReplyByPageRequest) (*post_grpc.CommentsResponse, error)   // 根据 CommentID 按页获取评论的回复
	CheckCommentDeleteAuth(context.Context, *post_grpc.CommentBelongRequest) (*post_grpc.BelongResponse, error)  // 用户是否有评论删除权限
	post_grpc.UnsafePostServiceServer
}
