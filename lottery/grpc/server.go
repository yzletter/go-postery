package grpc

import (
	"context"

	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	"github.com/yzletter/go-postery/lottery/dto"
	"github.com/yzletter/go-postery/lottery/service"
)

type LotteryServiceServer struct {
	svc service.LotteryService
	lottery_grpc.UnimplementedLotteryServiceServer
}

func NewLotteryServiceServer(svc service.LotteryService) *LotteryServiceServer {
	return &LotteryServiceServer{
		svc: svc,
	}
}

func (server *LotteryServiceServer) GetAllGifts(ctx context.Context, request *lottery_grpc.EmptyRequest) (*lottery_grpc.Gifts, error) {
	// 调用 Service
	gifts, err := server.svc.GetAllGifts(ctx)
	if err != nil {
		return &lottery_grpc.Gifts{}, err
	}

	respGifts := make([]*lottery_grpc.Gift, 0, len(gifts))
	for _, gift := range gifts {
		respGifts = append(respGifts, dto.ToGift(gift))
	}

	// 返回 Response
	return &lottery_grpc.Gifts{Gifts: respGifts}, nil
}

func (server *LotteryServiceServer) Lottery(ctx context.Context, id *lottery_grpc.UserID) (*lottery_grpc.Gift, error) {
	// 调用 Service
	gift, err := server.svc.Lottery(ctx, id.UserID)
	if err != nil {
		return &lottery_grpc.Gift{}, err
	}
	// 返回 Response
	return dto.ToGift(gift), nil
}

func (server *LotteryServiceServer) Pay(ctx context.Context, request *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error) {
	// 调用 Service
	if err := server.svc.Pay(ctx, request.UserID, request.GiftID); err != nil {
		return &lottery_grpc.EmptyResponse{}, err
	}
	// 返回 Response
	return &lottery_grpc.EmptyResponse{}, nil
}

func (server *LotteryServiceServer) GiveUp(ctx context.Context, request *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error) {
	// 调用 Service
	if err := server.svc.GiveUp(ctx, request.UserID, request.GiftID); err != nil {
		return &lottery_grpc.EmptyResponse{}, err
	}
	// 返回 Response
	return &lottery_grpc.EmptyResponse{}, nil
}

func (server *LotteryServiceServer) Result(ctx context.Context, id *lottery_grpc.UserID) (*lottery_grpc.Order, error) {
	// 调用 Service
	order, gift, err := server.svc.Result(ctx, id.UserID)
	if err != nil {
		return &lottery_grpc.Order{}, err
	}
	// 返回 Response
	return dto.ToOrder(order, gift), nil
}
