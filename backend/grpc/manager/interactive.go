package manager

import (
	"context"
	"log/slog"
	"time"

	interactive_grpc "github.com/yzletter/go-postery/api/proto/interactive/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

type InteractiveServiceManager struct {
	service string
	hub     ServiceHub
}

func NewInteractiveManager(service string, hub ServiceHub) *InteractiveServiceManager {
	return &InteractiveServiceManager{
		service: service,
		hub:     hub,
	}
}

func (manager *InteractiveServiceManager) GetPostInteractive(ctx context.Context, req *interactive_grpc.PostIDRequest) (*interactive_grpc.PostInteractive, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.PostInteractive
		resp, err = client.GetPostInteractive(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) GetUserInteractive(ctx context.Context, req *interactive_grpc.UserIDRequest) (*interactive_grpc.UserInteractive, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.UserInteractive
		resp, err = client.GetUserInteractive(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) Like(ctx context.Context, req *interactive_grpc.LikeRequest) (*interactive_grpc.InteractiveEmptyResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // 写入类调用只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.InteractiveEmptyResponse
		resp, err = client.Like(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) Unlike(ctx context.Context, req *interactive_grpc.LikeRequest) (*interactive_grpc.InteractiveEmptyResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // 写入类调用只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.InteractiveEmptyResponse
		resp, err = client.Unlike(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) CheckLike(ctx context.Context, req *interactive_grpc.LikeRequest) (*interactive_grpc.CheckLikeResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.CheckLikeResponse
		resp, err = client.CheckLike(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) Follow(ctx context.Context, req *interactive_grpc.FollowRequest) (*interactive_grpc.InteractiveEmptyResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // 写入类调用只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.InteractiveEmptyResponse
		resp, err = client.Follow(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) Unfollow(ctx context.Context, req *interactive_grpc.FollowRequest) (*interactive_grpc.InteractiveEmptyResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // 写入类调用只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.InteractiveEmptyResponse
		resp, err = client.Unfollow(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) IfFollow(ctx context.Context, req *interactive_grpc.FollowRequest) (*interactive_grpc.IfFollowResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.IfFollowResponse
		resp, err = client.IfFollow(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) Comment(ctx context.Context, req *interactive_grpc.CreateCommentRequest) (*interactive_grpc.InteractiveComment, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // 写入类调用只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.InteractiveComment
		resp, err = client.Comment(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) DelComment(ctx context.Context, req *interactive_grpc.DeleteCommentRequest) (*interactive_grpc.InteractiveEmptyResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 1                // 写入类调用只适合一次
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.InteractiveEmptyResponse
		resp, err = client.DelComment(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) ListCommentByPage(ctx context.Context, req *interactive_grpc.ListCommentByPageRequest) (*interactive_grpc.CommentsResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.CommentsResponse
		resp, err = client.ListCommentByPage(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) ListRepliesByPage(ctx context.Context, req *interactive_grpc.ListReplyByPageRequest) (*interactive_grpc.CommentsResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.CommentsResponse
		resp, err = client.ListRepliesByPage(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) CheckCommentDelAuth(ctx context.Context, req *interactive_grpc.CommentIDUserIDRequest) (*interactive_grpc.CheckCommentDelAuthResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.CheckCommentDelAuthResponse
		resp, err = client.CheckCommentDelAuth(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) GetFollowers(ctx context.Context, req *interactive_grpc.ListFollowRequest) (*interactive_grpc.ListFollowResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.ListFollowResponse
		resp, err = client.GetFollowers(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) GetFollowees(ctx context.Context, req *interactive_grpc.ListFollowRequest) (*interactive_grpc.ListFollowResponse, error) {
	var err = errs.ErrUnavailable // 暴露错误
	var tryCnt = 3                // 查询类调用可重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interactive_grpc.ListFollowResponse
		resp, err = client.GetFollowees(ctx, req) // 微服务调用
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err // 返回 grpc 错误
	}

	// 默认会返回服务调用失败
	return nil, err
}

func (manager *InteractiveServiceManager) StartHealthCheck(ctx context.Context) {
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

func (manager *InteractiveServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}

		client := interactive_grpc.NewInteractiveServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &interactive_grpc.HealthCheckRequest{}) // 健康探测
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
