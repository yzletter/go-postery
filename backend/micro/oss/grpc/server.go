package grpc

import (
	"context"

	oss_grpc "github.com/yzletter/go-postery/api/proto/oss/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/micro/oss/domain"
	"github.com/yzletter/go-postery/backend/micro/oss/service"
)

type OSSServiceServer struct {
	svc service.OSSService
	oss_grpc.UnimplementedOSSServiceServer
}

func NewOSSServiceServer(svc service.OSSService) *OSSServiceServer {
	return &OSSServiceServer{
		svc: svc,
	}
}

func (server *OSSServiceServer) SignUpload(ctx context.Context, req *oss_grpc.SignUploadRequest) (*oss_grpc.SignUploadResponse, error) {
	if req == nil || req.UserID <= 0 {
		return &oss_grpc.SignUploadResponse{}, errs.ErrInvalidArgument
	}

	resp, dir, err := server.svc.SignUpload(ctx, domain.UploadBiz(req.Biz), req.UserID, req.FileName)
	if err != nil {
		return &oss_grpc.SignUploadResponse{}, err
	}
	return &oss_grpc.SignUploadResponse{
		Response: resp,
		Dir:      dir,
	}, nil
}

func (server *OSSServiceServer) UploadCallback(ctx context.Context, req *oss_grpc.UploadCallbackRequest) (*oss_grpc.UploadCallbackResponse, error) {
	if req == nil || req.Bucket == "" || req.ObjectName == "" {
		return &oss_grpc.UploadCallbackResponse{}, errs.ErrInvalidArgument
	}

	result, err := server.svc.UploadCallback(ctx, req.Bucket, req.ObjectName, req.Size)
	if err != nil {
		return &oss_grpc.UploadCallbackResponse{}, err
	}
	return &oss_grpc.UploadCallbackResponse{
		Biz:        oss_grpc.UploadBiz(result.Biz),
		UserID:     result.UserID,
		ObjectName: result.ObjectName,
		SourceFile: result.SourceFile,
		Size:       result.Size,
	}, nil
}

func (server *OSSServiceServer) GetObjectURL(ctx context.Context, req *oss_grpc.GetObjectURLRequest) (*oss_grpc.GetObjectURLResponse, error) {
	if req == nil || req.ObjectName == "" {
		return &oss_grpc.GetObjectURLResponse{}, errs.ErrInvalidArgument
	}

	url, err := server.svc.GetObjectURL(ctx, req.ObjectName)
	if err != nil {
		return &oss_grpc.GetObjectURLResponse{}, err
	}
	return &oss_grpc.GetObjectURLResponse{URL: url}, nil
}

func (server *OSSServiceServer) HealthCheck(ctx context.Context, req *oss_grpc.HealthCheckRequest) (*oss_grpc.HealthCheckResponse, error) {
	return &oss_grpc.HealthCheckResponse{}, nil
}
