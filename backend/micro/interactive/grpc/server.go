package grpc

import (
	"context"

	interactive_grpc "github.com/yzletter/go-postery/api/proto/interactive/v1"
	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
	"github.com/yzletter/go-postery/backend/micro/interactive/service"
)

type InteractiveServiceServer struct {
	svc service.InteractiveService
	interactive_grpc.UnimplementedInteractiveServiceServer
}

func NewInteractiveServiceServer(svc service.InteractiveService) *InteractiveServiceServer {
	return &InteractiveServiceServer{
		svc: svc,
	}
}

func (server *InteractiveServiceServer) GetPostInteractive(ctx context.Context, req *interactive_grpc.PostIDRequest) (*interactive_grpc.PostInteractive, error) {
	postInter, err := server.svc.GetPostInteractive(ctx, req.PostID)
	if err != nil {
		return &interactive_grpc.PostInteractive{}, err
	}
	return &interactive_grpc.PostInteractive{
		ReadCnt:    postInter.ReadCnt,
		LikeCnt:    postInter.LikeCnt,
		CommentCnt: postInter.CommentCnt,
	}, nil
}

func (server *InteractiveServiceServer) GetUserInteractive(ctx context.Context, req *interactive_grpc.UserIDRequest) (*interactive_grpc.UserInteractive, error) {
	userInter, err := server.svc.GetUserInteractive(ctx, req.UserID)
	if err != nil {
		return &interactive_grpc.UserInteractive{}, err
	}
	return &interactive_grpc.UserInteractive{FollowCnt: userInter.FollowCnt}, nil
}

func (server *InteractiveServiceServer) Like(ctx context.Context, req *interactive_grpc.LikeRequest) (*interactive_grpc.InteractiveEmptyResponse, error) {
	if err := server.svc.Like(ctx, req.UserID, req.PostID); err != nil {
		return &interactive_grpc.InteractiveEmptyResponse{}, err
	}
	return &interactive_grpc.InteractiveEmptyResponse{}, nil
}

func (server *InteractiveServiceServer) Unlike(ctx context.Context, req *interactive_grpc.LikeRequest) (*interactive_grpc.InteractiveEmptyResponse, error) {
	if err := server.svc.Unlike(ctx, req.PostID, req.UserID); err != nil {
		return &interactive_grpc.InteractiveEmptyResponse{}, err
	}
	return &interactive_grpc.InteractiveEmptyResponse{}, nil
}

func (server *InteractiveServiceServer) CheckLike(ctx context.Context, req *interactive_grpc.LikeRequest) (*interactive_grpc.CheckLikeResponse, error) {
	result, err := server.svc.CheckLike(ctx, req.UserID, req.PostID)
	if err != nil {
		return &interactive_grpc.CheckLikeResponse{Result: false}, err
	}
	return &interactive_grpc.CheckLikeResponse{Result: result}, nil
}

func (server *InteractiveServiceServer) Follow(ctx context.Context, req *interactive_grpc.FollowRequest) (*interactive_grpc.InteractiveEmptyResponse, error) {
	if err := server.svc.Follow(ctx, req.FollowerID, req.FolloweeID); err != nil {
		return &interactive_grpc.InteractiveEmptyResponse{}, err
	}
	return &interactive_grpc.InteractiveEmptyResponse{}, nil
}

func (server *InteractiveServiceServer) Unfollow(ctx context.Context, req *interactive_grpc.FollowRequest) (*interactive_grpc.InteractiveEmptyResponse, error) {
	if err := server.svc.Unfollow(ctx, req.FollowerID, req.FolloweeID); err != nil {
		return &interactive_grpc.InteractiveEmptyResponse{}, err
	}
	return &interactive_grpc.InteractiveEmptyResponse{}, nil
}

func (server *InteractiveServiceServer) IfFollow(ctx context.Context, req *interactive_grpc.FollowRequest) (*interactive_grpc.IfFollowResponse, error) {
	result, err := server.svc.IfFollow(ctx, req.FollowerID, req.FolloweeID)
	if err != nil {
		return &interactive_grpc.IfFollowResponse{Result: -1}, err
	}
	return &interactive_grpc.IfFollowResponse{Result: int32(result)}, nil
}

func (server *InteractiveServiceServer) Comment(ctx context.Context, req *interactive_grpc.CreateCommentRequest) (*interactive_grpc.InteractiveComment, error) {
	comment, err := server.svc.Comment(ctx, req.PostID, req.ParentID, req.ReplyID, req.UserID, req.Content)
	if err != nil {
		return &interactive_grpc.InteractiveComment{}, err
	}
	return toComment(comment), nil
}

func (server *InteractiveServiceServer) DelComment(ctx context.Context, req *interactive_grpc.DeleteCommentRequest) (*interactive_grpc.InteractiveEmptyResponse, error) {
	if err := server.svc.DelComment(ctx, req.CommentID, req.UserID); err != nil {
		return &interactive_grpc.InteractiveEmptyResponse{}, err
	}
	return &interactive_grpc.InteractiveEmptyResponse{}, nil
}

func (server *InteractiveServiceServer) ListCommentByPage(ctx context.Context, req *interactive_grpc.ListCommentByPageRequest) (*interactive_grpc.CommentsResponse, error) {
	total, comments, err := server.svc.ListCommentByPage(ctx, req.PostID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &interactive_grpc.CommentsResponse{}, err
	}

	respComments := make([]*interactive_grpc.InteractiveComment, 0, len(comments))
	for _, comment := range comments {
		respComments = append(respComments, toComment(comment))
	}

	return &interactive_grpc.CommentsResponse{
		Count:    uint64(total),
		Comments: respComments,
	}, nil
}

func (server *InteractiveServiceServer) ListRepliesByPage(ctx context.Context, req *interactive_grpc.ListReplyByPageRequest) (*interactive_grpc.CommentsResponse, error) {
	total, comments, err := server.svc.ListRepliesByPage(ctx, req.CommentID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &interactive_grpc.CommentsResponse{}, err
	}

	respComments := make([]*interactive_grpc.InteractiveComment, 0, len(comments))
	for _, comment := range comments {
		respComments = append(respComments, toComment(comment))
	}

	return &interactive_grpc.CommentsResponse{
		Count:    uint64(total),
		Comments: respComments,
	}, nil
}

func (server *InteractiveServiceServer) CheckCommentDelAuth(ctx context.Context, req *interactive_grpc.CommentIDUserIDRequest) (*interactive_grpc.CheckCommentDelAuthResponse, error) {
	result, err := server.svc.CheckCommentDelAuth(ctx, req.CommentID, req.UserID)
	if err != nil {
		return &interactive_grpc.CheckCommentDelAuthResponse{Result: false}, err
	}
	return &interactive_grpc.CheckCommentDelAuthResponse{Result: result}, nil
}

func (server *InteractiveServiceServer) GetFollowers(ctx context.Context, req *interactive_grpc.ListFollowRequest) (*interactive_grpc.ListFollowResponse, error) {
	total, IDs, err := server.svc.GetFollowers(ctx, req.UserID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &interactive_grpc.ListFollowResponse{}, err
	}
	return &interactive_grpc.ListFollowResponse{Count: uint64(total), IDs: IDs}, nil
}

func (server *InteractiveServiceServer) GetFollowees(ctx context.Context, req *interactive_grpc.ListFollowRequest) (*interactive_grpc.ListFollowResponse, error) {
	total, IDs, err := server.svc.GetFollowees(ctx, req.UserID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return &interactive_grpc.ListFollowResponse{}, err
	}
	return &interactive_grpc.ListFollowResponse{Count: uint64(total), IDs: IDs}, nil
}

func (server *InteractiveServiceServer) HealthCheck(ctx context.Context, req *interactive_grpc.HealthCheckRequest) (*interactive_grpc.HealthCheckResponse, error) {
	return &interactive_grpc.HealthCheckResponse{}, nil
}

func toComment(comment domain.Comment) *interactive_grpc.InteractiveComment {
	return &interactive_grpc.InteractiveComment{
		ID:       comment.ID,
		PostID:   comment.PostID,
		ParentID: comment.ParentID,
		ReplyID:  comment.ReplyID,
		UserID:   comment.UserID,
		Content:  comment.Content,
	}
}
