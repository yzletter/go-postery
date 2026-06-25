package grpc

import (
	"context"
	"errors"
	"time"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/micro/post/domain"
	"github.com/yzletter/go-postery/backend/micro/post/grpc/dto"
	"github.com/yzletter/go-postery/backend/micro/post/service"
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
	post, err := server.svc.Create(ctx, domain.Post{
		User:        domain.User{UserID: req.UserID},
		Title:       req.Title,
		Content:     req.Content,
		ContentType: int(req.ContentType),
		Tags:        req.Tags,
	})
	if err != nil {
		return &post_grpc.PostDetail{}, err
	}
	return dto.ToPostDetail(post), nil
}

func (server *PostServiceServer) GetDetailByID(ctx context.Context, req *post_grpc.GetDetailByIDRequest) (*post_grpc.PostDetail, error) {
	post, err := server.svc.GetDetailByID(ctx, req.PostID, req.AddViewCnt)
	if err != nil {
		return &post_grpc.PostDetail{}, err
	}
	return dto.ToPostDetail(post), nil
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

func (server *PostServiceServer) GetPostByTime(ctx context.Context, req *post_grpc.GetPostByTimeRequest) (*post_grpc.PostIDs, error) {
	timeAt, err := time.Parse(time.RFC3339, req.TimeAt)
	if err != nil {
		return &post_grpc.PostIDs{}, errs.ErrInvalidArgument
	}

	ids, err := server.svc.GetPostByTime(ctx, timeAt)
	if err != nil {
		return &post_grpc.PostIDs{}, err
	}
	return &post_grpc.PostIDs{IDs: ids}, nil
}

func (server *PostServiceServer) Update(ctx context.Context, req *post_grpc.UpdateRequest) (*post_grpc.PostEmptyResponse, error) {
	if err := server.svc.Update(ctx, domain.Post{
		ID:      req.PostID,
		User:    domain.User{UserID: req.UserID},
		Title:   req.Title,
		Content: req.Content,
		Tags:    req.Tags,
	}); err != nil {
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
		postDetails = append(postDetails, dto.ToPostDetail(post))
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
		postBriefs = append(postBriefs, dto.ToPostBrief(post.Briefed()))
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
		postDetails = append(postDetails, dto.ToPostDetail(post))
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

func (server *PostServiceServer) ExistPost(ctx context.Context, req *post_grpc.ExistPostRequest) (*post_grpc.ExistPostResponse, error) {
	_, err := server.svc.GetBriefByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return &post_grpc.ExistPostResponse{Exist: false}, nil
		}
		return &post_grpc.ExistPostResponse{Exist: false}, err
	}
	return &post_grpc.ExistPostResponse{Exist: true}, nil
}

func (server *PostServiceServer) CheckPostAuth(ctx context.Context, req *post_grpc.CheckPostAuthRequest) (*post_grpc.CheckPostAuthResponse, error) {
	result, err := server.svc.Belong(ctx, req.UserID, req.PostID)
	if err != nil {
		return &post_grpc.CheckPostAuthResponse{Exist: false}, err
	}
	return &post_grpc.CheckPostAuthResponse{Exist: result}, nil
}

func (server *PostServiceServer) HealthCheck(ctx context.Context, req *post_grpc.HealthCheckRequest) (*post_grpc.HealthCheckResponse, error) {
	return &post_grpc.HealthCheckResponse{}, nil
}
