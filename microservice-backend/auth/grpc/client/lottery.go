package client

import (
	"context"
	"time"

	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type lotteryClient struct {
	conn   *grpc.ClientConn
	client lottery_grpc.LotteryServiceClient
}

func NewLotteryClient() (LotteryClient, error) {
	// 建议：启用 ka，避免中间网络设备把长连接静默掐掉
	ka := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(
		LotteryClientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 生产用 TLS
		CircuitBreakerDialOption(),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()), // Jaeger
		grpc.WithKeepaliveParams(ka),
	)
	if err != nil {
		return nil, err
	}

	return &lotteryClient{
		conn:   conn,
		client: lottery_grpc.NewLotteryServiceClient(conn),
	}, nil
}

func (client *lotteryClient) Close() {
	_ = client.conn.Close()
}

func (client *lotteryClient) GetAllGifts(ctx context.Context, req *lottery_grpc.EmptyRequest) (*lottery_grpc.Gifts, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.GetAllGifts(ctx, req)
}

func (client *lotteryClient) Lottery(ctx context.Context, req *lottery_grpc.UserID) (*lottery_grpc.LotteryResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Lottery(ctx, req)
}

func (client *lotteryClient) Pay(ctx context.Context, req *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Pay(ctx, req)
}

func (client *lotteryClient) GiveUp(ctx context.Context, req *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.GiveUp(ctx, req)
}

func (client *lotteryClient) Result(ctx context.Context, req *lottery_grpc.UserID) (*lottery_grpc.Order, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 8000*time.Millisecond)
	defer cancel()

	return client.client.Result(ctx, req)
}
