package server

import (
	"context"

	lottery_grpc2 "github.com/yzletter/go-postery/backend/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/backend/micro/lottery/dto"
	"github.com/yzletter/go-postery/backend/micro/lottery/service"
)

type LotteryServiceServer struct {
	svc service.LotteryService
	lottery_grpc2.UnimplementedLotteryServiceServer
}

func NewLotteryServiceServer(svc service.LotteryService) *LotteryServiceServer {
	return &LotteryServiceServer{
		svc: svc,
	}
}

func (server *LotteryServiceServer) GetAllGifts(ctx context.Context, request *lottery_grpc2.EmptyRequest) (*lottery_grpc2.Gifts, error) {
	// 调用 Service
	gifts, err := server.svc.GetAllGifts(ctx)
	if err != nil {
		return &lottery_grpc2.Gifts{}, err
	}

	respGifts := make([]*lottery_grpc2.Gift, 0, len(gifts))
	for _, gift := range gifts {
		respGifts = append(respGifts, dto.ToGift(gift))
	}

	// 返回 Response
	return &lottery_grpc2.Gifts{Gifts: respGifts}, nil
}

func (server *LotteryServiceServer) Lottery(ctx context.Context, id *lottery_grpc2.UserID) (*lottery_grpc2.LotteryResponse, error) {
	// 调用 Service
	result, err := server.svc.Lottery(ctx, id.UserID)
	if err != nil {
		return &lottery_grpc2.LotteryResponse{}, err
	}
	// 返回 Response
	return dto.ToLotteryResponse(result), nil
}

func (server *LotteryServiceServer) Pay(ctx context.Context, request *lottery_grpc2.LotteryCommonRequest) (*lottery_grpc2.EmptyResponse, error) {
	// 调用 Service
	if err := server.svc.Pay(ctx, request.UserID, request.TempOrderID, request.GiftID); err != nil {
		return &lottery_grpc2.EmptyResponse{}, err
	}
	// 返回 Response
	return &lottery_grpc2.EmptyResponse{}, nil
}

func (server *LotteryServiceServer) GiveUp(ctx context.Context, request *lottery_grpc2.LotteryCommonRequest) (*lottery_grpc2.EmptyResponse, error) {
	// 调用 Service
	if err := server.svc.GiveUp(ctx, request.UserID, request.TempOrderID, request.GiftID); err != nil {
		return &lottery_grpc2.EmptyResponse{}, err
	}
	// 返回 Response
	return &lottery_grpc2.EmptyResponse{}, nil
}

func (server *LotteryServiceServer) Result(ctx context.Context, id *lottery_grpc2.UserID) (*lottery_grpc2.Order, error) {
	// 调用 Service
	order, gift, err := server.svc.Result(ctx, id.UserID)
	if err != nil {
		return &lottery_grpc2.Order{}, err
	}
	// 返回 Response
	return dto.ToOrder(order, gift), nil
}

func (server *LotteryServiceServer) HealthCheck(ctx context.Context, request *lottery_grpc2.HealthCheckRequest) (*lottery_grpc2.HealthCheckResponse, error) {
	return &lottery_grpc2.HealthCheckResponse{}, nil
}
