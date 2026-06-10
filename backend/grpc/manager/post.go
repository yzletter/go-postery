package manager

import (
	"context"
	"log/slog"
	"time"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"github.com/yzletter/go-postery/backend/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostServiceManager struct {
	service string
	hub     ServiceHub
}

func NewPostManager(service string, hub ServiceHub) *PostServiceManager {
	return &PostServiceManager{
		service: service,
		hub:     hub,
	}
}

func (manager *PostServiceManager) Create(ctx context.Context, req *post_grpc.CreatePostRequest) (*post_grpc.PostDetail, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostDetail
		resp, err = client.Create(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) GetDetailByID(ctx context.Context, req *post_grpc.GetDetailByIDRequest) (*post_grpc.PostDetail, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostDetail
		resp, err = client.GetDetailByID(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) GetBriefByID(ctx context.Context, req *post_grpc.GetBriefByIDRequest) (*post_grpc.PostBrief, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostBrief
		resp, err = client.GetBriefByID(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) Top(ctx context.Context, req *post_grpc.PostEmptyRequest) (*post_grpc.TopResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.TopResponse
		resp, err = client.Top(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) Update(ctx context.Context, req *post_grpc.UpdateRequest) (*post_grpc.PostEmptyResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostEmptyResponse
		resp, err = client.Update(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) ListByPage(ctx context.Context, req *post_grpc.ListByPageRequest) (*post_grpc.PostDetailsResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostDetailsResponse
		resp, err = client.ListByPage(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) ListByPageAndUid(ctx context.Context, req *post_grpc.ListByPageAndUidRequest) (*post_grpc.PostBriefsResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostBriefsResponse
		resp, err = client.ListByPageAndUid(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) ListByPageAndTag(ctx context.Context, req *post_grpc.ListByPageAndTagRequest) (*post_grpc.PostDetailsResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostDetailsResponse
		resp, err = client.ListByPageAndTag(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) Belong(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.BelongResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.BelongResponse
		resp, err = client.Belong(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) Delete(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostEmptyResponse
		resp, err = client.Delete(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) Like(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostEmptyResponse
		resp, err = client.Like(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) Unlike(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostEmptyResponse
		resp, err = client.Unlike(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) IfLike(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.IfLikeResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.IfLikeResponse
		resp, err = client.IfLike(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) CreateComment(ctx context.Context, req *post_grpc.CreateCommentRequest) (*post_grpc.Comment, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.Comment
		resp, err = client.CreateComment(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) DeleteComment(ctx context.Context, req *post_grpc.DeleteCommentRequest) (*post_grpc.PostEmptyResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.PostEmptyResponse
		resp, err = client.DeleteComment(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) ListCommentByPage(ctx context.Context, req *post_grpc.ListCommentByPageRequest) (*post_grpc.CommentsResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.CommentsResponse
		resp, err = client.ListCommentByPage(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) ListRepliesByPage(ctx context.Context, req *post_grpc.ListReplyByPageRequest) (*post_grpc.CommentsResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.CommentsResponse
		resp, err = client.ListRepliesByPage(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) CheckCommentDeleteAuth(ctx context.Context, req *post_grpc.CommentBelongRequest) (*post_grpc.BelongResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := post_grpc.NewPostServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *post_grpc.BelongResponse
		resp, err = client.CheckCommentDeleteAuth(ctx, req)
		cancel()

		if err != nil && status.Code(err) == codes.Internal {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *PostServiceManager) StartHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			manager.checkOnce(ctx)
		}
	}
}

func (manager *PostServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}

		client := post_grpc.NewPostServiceClient(endpoint.Conn)
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &post_grpc.HealthCheckRequest{})
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
