package manager

import (
	"context"
	"log/slog"
	"time"

	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/backend/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LotteryServiceManager struct {
	service string
	hub     ServiceHub
}

func NewLotteryManager(service string, hub ServiceHub) *LotteryServiceManager {
	return &LotteryServiceManager{
		service: service,
		hub:     hub,
	}
}

func (manager *LotteryServiceManager) GetAllGifts(ctx context.Context, req *lottery_grpc.EmptyRequest) (*lottery_grpc.Gifts, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := lottery_grpc.NewLotteryServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *lottery_grpc.Gifts
		resp, err = client.GetAllGifts(ctx, req)
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

func (manager *LotteryServiceManager) Lottery(ctx context.Context, req *lottery_grpc.UserID) (*lottery_grpc.LotteryResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1 // 抽奖有副作用, 不重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := lottery_grpc.NewLotteryServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *lottery_grpc.LotteryResponse
		resp, err = client.Lottery(ctx, req)
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

func (manager *LotteryServiceManager) Pay(ctx context.Context, req *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1 // 支付有副作用, 不重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := lottery_grpc.NewLotteryServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *lottery_grpc.EmptyResponse
		resp, err = client.Pay(ctx, req)
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

func (manager *LotteryServiceManager) GiveUp(ctx context.Context, req *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1 // 放弃奖品有副作用, 不重试
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := lottery_grpc.NewLotteryServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *lottery_grpc.EmptyResponse
		resp, err = client.GiveUp(ctx, req)
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

func (manager *LotteryServiceManager) Result(ctx context.Context, req *lottery_grpc.UserID) (*lottery_grpc.Order, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 3
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}
		client := lottery_grpc.NewLotteryServiceClient(endpoint.Conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *lottery_grpc.Order
		resp, err = client.Result(ctx, req)
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

func (manager *LotteryServiceManager) StartHealthCheck(ctx context.Context) {
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

func (manager *LotteryServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil || endpoint.Conn == nil {
			continue
		}

		client := lottery_grpc.NewLotteryServiceClient(endpoint.Conn)
		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &lottery_grpc.HealthCheckRequest{})
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
