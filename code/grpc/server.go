package grpc_server

import (
	"context"

	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	"github.com/yzletter/go-postery/code/service"
)

// CodeServiceServer gRPC 服务端
type CodeServiceServer struct {
	svc service.CodeService
	code_grpc.UnimplementedCodeServiceServer
}

// Send 发送验证码
func (server *CodeServiceServer) Send(ctx context.Context, req *code_grpc.SendCodeRequest) (*code_grpc.SendCodeResponse, error) {
	// 调用 Service
	if err := server.svc.Send(ctx, int(req.Biz), req.Identifier); err != nil {
		return &code_grpc.SendCodeResponse{}, err
	}
	// 返回 Response
	return &code_grpc.SendCodeResponse{}, nil
}

// Verify 校验验证码
func (server *CodeServiceServer) Verify(ctx context.Context, req *code_grpc.CheckCodeRequest) (*code_grpc.CheckCodeResponse, error) {
	// 调用 Service
	if res, err := server.svc.Verify(ctx, int(req.Biz), req.Identifier, req.Code); err != nil {
		return &code_grpc.CheckCodeResponse{Result: false}, err
	} else {
		return &code_grpc.CheckCodeResponse{Result: res}, nil
	}
}

func NewCodeServiceServer(svc service.CodeService) *CodeServiceServer {
	return &CodeServiceServer{
		svc: svc,
	}
}
