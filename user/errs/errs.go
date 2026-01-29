package errs

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidArgument = status.Error(codes.InvalidArgument, "请求参数错误")
	ErrInternal        = status.Error(codes.Internal, "系统繁忙，请稍后重试")
	ErrNotFound        = status.Error(codes.NotFound, "资源不存在")
	ErrAlreadyExits    = status.Error(codes.AlreadyExists, "重复创建")
)
