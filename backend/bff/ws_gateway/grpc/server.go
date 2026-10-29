package server

import (
	"context"
	"encoding/json"

	"github.com/bytedance/sonic"
	ws_gateway_grpc2 "github.com/yzletter/go-postery/backend/api/proto/ws_gateway/v1"
	"github.com/yzletter/go-postery/backend/bff/ws_gateway/domain"
	"github.com/yzletter/go-postery/backend/bff/ws_gateway/service"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

type WSGatewayServiceServer struct {
	gateway *service.WebsocketGateway
	ws_gateway_grpc2.UnimplementedWSGatewayServiceServer
}

func NewWSGatewayServiceServer(gateway *service.WebsocketGateway) *WSGatewayServiceServer {
	return &WSGatewayServiceServer{
		gateway: gateway,
	}
}

func (server *WSGatewayServiceServer) Push(ctx context.Context, req *ws_gateway_grpc2.PushRequest) (*ws_gateway_grpc2.PushResponse, error) {
	if req.UserID == 0 || req.BizType == "" {
		return &ws_gateway_grpc2.PushResponse{}, errs.ErrInvalidArgument
	}
	if server.gateway == nil {
		return &ws_gateway_grpc2.PushResponse{}, errs.ErrUnavailable
	}

	var bizData any
	if len(req.BizData) > 0 {
		bizData = json.RawMessage(req.BizData)
	}
	data, err := sonic.Marshal(service.WSMessage{
		BizType: req.BizType,
		BizData: bizData,
	})
	if err != nil {
		return &ws_gateway_grpc2.PushResponse{}, err
	}
	if err = server.gateway.Push(ctx, req.UserID, convert(req.ConnBiz), data); err != nil {
		return &ws_gateway_grpc2.PushResponse{}, errs.ErrNotFound
	}
	return &ws_gateway_grpc2.PushResponse{}, nil
}

func (server *WSGatewayServiceServer) HealthCheck(ctx context.Context, req *ws_gateway_grpc2.HealthCheckRequest) (*ws_gateway_grpc2.HealthCheckResponse, error) {
	return &ws_gateway_grpc2.HealthCheckResponse{}, nil
}

// 将 grpc api 的常量转为 dto
func convert(biz ws_gateway_grpc2.ConnBiz) domain.ConnType {
	switch biz {
	case ws_gateway_grpc2.ConnBiz_CONNCECTION_Biz_Interview:
		return domain.BizInterview
	case ws_gateway_grpc2.ConnBiz_CONNCECTION_Biz_Session:
		return domain.BizSession
	}
	return ""
}
