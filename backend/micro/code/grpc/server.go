package server

import (
	"context"

	code_grpc2 "github.com/yzletter/go-postery/backend/api/proto/code/v1"
	"github.com/yzletter/go-postery/backend/micro/code/domain"
	"github.com/yzletter/go-postery/backend/micro/code/service"
)

// CodeServiceServer gRPC 服务端
type CodeServiceServer struct {
	svc service.CodeService
	code_grpc2.UnimplementedCodeServiceServer
}

func NewCodeServiceServer(svc service.CodeService) *CodeServiceServer {
	return &CodeServiceServer{
		svc: svc,
	}
}

// Send 发送验证码
func (server *CodeServiceServer) Send(ctx context.Context, req *code_grpc2.SendCodeRequest) (*code_grpc2.SendCodeResponse, error) {
	// 调用 Service
	if err := server.svc.Send(ctx, convert(req.Biz), req.Identifier); err != nil {
		return &code_grpc2.SendCodeResponse{}, err
	}
	// 返回 Response
	return &code_grpc2.SendCodeResponse{}, nil
}

// Verify 校验验证码
func (server *CodeServiceServer) Verify(ctx context.Context, req *code_grpc2.VerifyCodeRequest) (*code_grpc2.VerifyCodeResponse, error) {
	// 调用 Service
	if res, err := server.svc.Verify(ctx, convert(req.Biz), req.Identifier, req.Code); err != nil {
		return &code_grpc2.VerifyCodeResponse{Result: false}, err
	} else {
		return &code_grpc2.VerifyCodeResponse{Result: res}, nil
	}
}

func (server *CodeServiceServer) HealthCheck(ctx context.Context, req *code_grpc2.HealthCheckRequest) (*code_grpc2.HealthCheckResponse, error) {
	return &code_grpc2.HealthCheckResponse{}, nil
}

func convert(biz code_grpc2.CodeBiz) domain.BizType {
	switch biz {
	case code_grpc2.CodeBiz_CODE_BIZ_SMS:
		return domain.BizSMS
	case code_grpc2.CodeBiz_CODE_BIZ_Email:
		return domain.BizEmail
	}
	return 0
}
