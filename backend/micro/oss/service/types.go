package service

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/oss/domain"
)

type OSSService interface {
	// SignUpload 获取 OSS 上传签名
	SignUpload(ctx context.Context, biz domain.UploadBiz, userID int64, fileName string) (string, string, error)

	// UploadCallback 解析 OSS 上传回调对象
	UploadCallback(ctx context.Context, bucket string, objectName string, size string) (domain.CallbackResult, error)

	// GetObjectURL 获取 OSS 对象下载预签名 URL
	GetObjectURL(ctx context.Context, objectName string) (string, error)
}
