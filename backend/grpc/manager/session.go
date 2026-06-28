package manager

import (
	"context"
	"log/slog"
	"time"

	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

type SessionServiceManager struct {
	service string
	hub     ServiceHub
}

func NewSessionManager(service string, hub ServiceHub) *SessionServiceManager {
	return &SessionServiceManager{
		service: service,
		hub:     hub,
	}
}

func (manager *SessionServiceManager) ListByUID(ctx context.Context, req *session_grpc.UserID) (*session_grpc.Sessions, error) {
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
		client := session_grpc.NewSessionServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *session_grpc.Sessions
		resp, err = client.ListByUID(ctx, req) // 微服务调用
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

func (manager *SessionServiceManager) GetSession(ctx context.Context, req *session_grpc.BothUserID) (*session_grpc.Session, error) {
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
		client := session_grpc.NewSessionServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *session_grpc.Session
		resp, err = client.GetSession(ctx, req) // 微服务调用
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

func (manager *SessionServiceManager) GetHistoryMessagesByPage(ctx context.Context, req *session_grpc.GetHistoryMessagesByPageRequest) (*session_grpc.GetHistoryMessagesByPageResponse, error) {
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
		client := session_grpc.NewSessionServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *session_grpc.GetHistoryMessagesByPageResponse
		resp, err = client.GetHistoryMessagesByPage(ctx, req) // 微服务调用
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

func (manager *SessionServiceManager) Delete(ctx context.Context, req *session_grpc.DeleteRequest) (*session_grpc.SessionEmptyResponse, error) {
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
		client := session_grpc.NewSessionServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *session_grpc.SessionEmptyResponse
		resp, err = client.Delete(ctx, req) // 微服务调用
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

func (manager *SessionServiceManager) UpdateUnread(ctx context.Context, req *session_grpc.UpdateUnreadRequest) (*session_grpc.SessionEmptyResponse, error) {
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
		client := session_grpc.NewSessionServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *session_grpc.SessionEmptyResponse
		resp, err = client.UpdateUnread(ctx, req) // 微服务调用
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

func (manager *SessionServiceManager) ClearUnread(ctx context.Context, req *session_grpc.ClearUnreadRequest) (*session_grpc.SessionEmptyResponse, error) {
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
		client := session_grpc.NewSessionServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *session_grpc.SessionEmptyResponse
		resp, err = client.ClearUnread(ctx, req) // 微服务调用
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

func (manager *SessionServiceManager) CreateMessage(ctx context.Context, req *session_grpc.Message) (*session_grpc.Message, error) {
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
		client := session_grpc.NewSessionServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *session_grpc.Message
		resp, err = client.CreateMessage(ctx, req) // 微服务调用
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

func (manager *SessionServiceManager) StartHealthCheck(ctx context.Context) {
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

func (manager *SessionServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}

		client := session_grpc.NewSessionServiceClient(conn) // 建立 Client

		// 添加超时控制
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &session_grpc.HealthCheckRequest{}) // 健康探测
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
