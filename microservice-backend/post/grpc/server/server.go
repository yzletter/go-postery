package server

import (
	"context"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"github.com/yzletter/go-postery/microservice-backend/post/dto"
	"github.com/yzletter/go-postery/microservice-backend/post/service"
)

type PostServiceServer struct {
	svc service.PostService
	post_grpc.UnimplementedPostServiceServer
}

func NewPostServiceServer(svc service.PostService) *PostServiceServer {
	return &PostServiceServer{
		svc: svc,
	}
}

func (server *PostServiceServer) Create(ctx context.Context, req *post_grpc.CreatePostRequest) (*post_grpc.PostDetail, error) {
	postWithTags, err := server.svc.Create(ctx, req.UserID, req.Title, req.Content, int(req.ContentType), req.Tags)
	if err != nil {
		return &post_grpc.PostDetail{}, err
	}
	return dto.ToPostDetail(postWithTags.Post, postWithTags.Tags), nil
}

func (server *PostServiceServer) GetDetailByID(ctx context.Context, req *post_grpc.GetDetailByIDRequest) (*post_grpc.PostDetail, error) {
	postWithTags, err := server.svc.GetDetailByID(ctx, req.PostID, req.AddViewCnt)
	if err != nil {
		return &post_grpc.PostDetail{}, err
	}
	return dto.ToPostDetail(postWithTags.Post, postWithTags.Tags), nil
}

func (server *PostServiceServer) GetBriefByID(ctx context.Context, req *post_grpc.GetBriefByIDRequest) (*post_grpc.PostBrief, error) {
	post, err := server.svc.GetBriefByID(ctx, req.PostID)
	if err != nil {
		return &post_grpc.PostBrief{}, err
	}
	return dto.ToPostBrief(post), nil
}

func (server *PostServiceServer) Top(ctx context.Context, req *post_grpc.PostEmptyRequest) (*post_grpc.TopResponse, error) {
	_ = req
	posts, scores, err := server.svc.Top(ctx)
	if err != nil {
		return &post_grpc.TopResponse{}, err
	}

	topPosts := make([]*post_grpc.TopPost, 0, len(posts))
	for i, post := range posts {
		topPosts = append(topPosts, dto.ToTopPost(post, scores[i]))
	}

	return &post_grpc.TopResponse{TopPosts: topPosts}, nil
}

func (server *PostServiceServer) Update(ctx context.Context, req *post_grpc.UpdateRequest) (*post_grpc.PostEmptyResponse, error) {
	if err := server.svc.Update(ctx, req.UserID, req.PostID, req.Title, req.Content, req.Tags); err != nil {
		return &post_grpc.PostEmptyResponse{}, err
	}
	return &post_grpc.PostEmptyResponse{}, nil
}

func (server *PostServiceServer) ListByPage(ctx context.Context, req *post_grpc.ListByPageRequest) (*post_grpc.PostDetailsResponse, error) {
	total, posts, err := server.svc.ListByPage(ctx, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &post_grpc.PostDetailsResponse{}, err
	}

	postDetails := make([]*post_grpc.PostDetail, 0, len(posts))
	for _, post := range posts {
		postDetails = append(postDetails, dto.ToPostDetail(post.Post, post.Tags))
	}

	return &post_grpc.PostDetailsResponse{
		Count:       uint64(total),
		PostDetails: postDetails,
	}, nil
}

func (server *PostServiceServer) ListByPageAndUid(ctx context.Context, req *post_grpc.ListByPageAndUidRequest) (*post_grpc.PostBriefsResponse, error) {
	total, posts, err := server.svc.ListByPageAndUid(ctx, req.UserID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &post_grpc.PostBriefsResponse{}, err
	}

	postBriefs := make([]*post_grpc.PostBrief, 0, len(posts))
	for _, post := range posts {
		postBriefs = append(postBriefs, dto.ToPostBrief(post))
	}

	return &post_grpc.PostBriefsResponse{
		Count:      uint64(total),
		PostBriefs: postBriefs,
	}, nil
}

func (server *PostServiceServer) ListByPageAndTag(ctx context.Context, req *post_grpc.ListByPageAndTagRequest) (*post_grpc.PostDetailsResponse, error) {
	total, posts, err := server.svc.ListByPageAndTag(ctx, req.Tag, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &post_grpc.PostDetailsResponse{}, err
	}

	postDetails := make([]*post_grpc.PostDetail, 0, len(posts))
	for _, post := range posts {
		postDetails = append(postDetails, dto.ToPostDetail(post.Post, post.Tags))
	}

	return &post_grpc.PostDetailsResponse{
		Count:       uint64(total),
		PostDetails: postDetails,
	}, nil
}

func (server *PostServiceServer) Belong(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.BelongResponse, error) {
	result, err := server.svc.Belong(ctx, req.UserID, req.PostID)
	if err != nil {
		return &post_grpc.BelongResponse{Result: false}, err
	}
	return &post_grpc.BelongResponse{Result: result}, nil
}

func (server *PostServiceServer) Delete(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error) {
	if err := server.svc.Delete(ctx, req.UserID, req.PostID); err != nil {
		return &post_grpc.PostEmptyResponse{}, err
	}
	return &post_grpc.PostEmptyResponse{}, nil
}

func (server *PostServiceServer) Like(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error) {
	if err := server.svc.Like(ctx, req.UserID, req.PostID); err != nil {
		return &post_grpc.PostEmptyResponse{}, err
	}
	return &post_grpc.PostEmptyResponse{}, nil
}

func (server *PostServiceServer) Unlike(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error) {
	if err := server.svc.Unlike(ctx, req.UserID, req.PostID); err != nil {
		return &post_grpc.PostEmptyResponse{}, err
	}
	return &post_grpc.PostEmptyResponse{}, nil
}

func (server *PostServiceServer) IfLike(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.IfLikeResponse, error) {
	result, err := server.svc.IfLike(ctx, req.UserID, req.PostID)
	if err != nil {
		return &post_grpc.IfLikeResponse{Result: false}, err
	}
	return &post_grpc.IfLikeResponse{Result: result}, nil
}

func (server *PostServiceServer) CreateComment(ctx context.Context, req *post_grpc.CreateCommentRequest) (*post_grpc.Comment, error) {
	comment, err := server.svc.CreateComment(ctx, req.PostID, req.ParentID, req.ReplyID, req.UserID, req.Content)
	if err != nil {
		return &post_grpc.Comment{}, err
	}
	return dto.ToComment(comment), nil
}

func (server *PostServiceServer) DeleteComment(ctx context.Context, req *post_grpc.DeleteCommentRequest) (*post_grpc.PostEmptyResponse, error) {
	if err := server.svc.DeleteComment(ctx, req.CommentID, req.UserID); err != nil {
		return &post_grpc.PostEmptyResponse{}, err
	}
	return &post_grpc.PostEmptyResponse{}, nil
}

func (server *PostServiceServer) ListCommentByPage(ctx context.Context, req *post_grpc.ListCommentByPageRequest) (*post_grpc.CommentsResponse, error) {
	total, comments, err := server.svc.ListCommentByPage(ctx, req.PostID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &post_grpc.CommentsResponse{}, err
	}

	respComments := make([]*post_grpc.Comment, 0, len(comments))
	for _, comment := range comments {
		respComments = append(respComments, dto.ToComment(comment))
	}

	return &post_grpc.CommentsResponse{
		Count:    uint64(total),
		Comments: respComments,
	}, nil
}

func (server *PostServiceServer) ListRepliesByPage(ctx context.Context, req *post_grpc.ListReplyByPageRequest) (*post_grpc.CommentsResponse, error) {
	total, comments, err := server.svc.ListRepliesByPage(ctx, req.CommentID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &post_grpc.CommentsResponse{}, err
	}

	respComments := make([]*post_grpc.Comment, 0, len(comments))
	for _, comment := range comments {
		respComments = append(respComments, dto.ToComment(comment))
	}

	return &post_grpc.CommentsResponse{
		Count:    uint64(total),
		Comments: respComments,
	}, nil
}

func (server *PostServiceServer) CheckCommentDeleteAuth(ctx context.Context, req *post_grpc.CommentBelongRequest) (*post_grpc.BelongResponse, error) {
	result, err := server.svc.CheckCommentDeleteAuth(ctx, req.CommentID, req.UserID)
	if err != nil {
		return &post_grpc.BelongResponse{Result: false}, err
	}
	return &post_grpc.BelongResponse{Result: result}, nil
}
