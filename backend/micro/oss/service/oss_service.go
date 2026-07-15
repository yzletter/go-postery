package service

import (
	"context"
	"log/slog"
	"path"
	"strconv"
	"strings"

	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/micro/oss/domain"
	"github.com/yzletter/go-postery/backend/ports"
)

const defaultBucket = "go-postery"

type ossService struct {
	ossManager  ports.OSSManager
	bucket      string
	callbackURL string
}

func NewOSSService(ossManager ports.OSSManager, config conf.OSSConfig) OSSService {
	bucket := config.Bucket
	if bucket == "" {
		bucket = defaultBucket
	}

	return &ossService{
		ossManager:  ossManager,
		bucket:      bucket,
		callbackURL: config.CallbackURL,
	}
}

func serviceLogger(method string) *slog.Logger {
	return slog.With("component", "oss_service", "method", method)
}

func (svc *ossService) SignUpload(ctx context.Context, biz domain.UploadBiz, userID int64, fileName string) (string, string, error) {
	logger := serviceLogger("SignUpload").With("biz", biz, "user_id", userID, "file_name", fileName)
	if userID <= 0 {
		logger.Debug("sign upload rejected: invalid user id")
		return "", "", errs.ErrInvalidArgument
	}

	dir, err := uploadDir(biz, userID)
	if err != nil {
		logger.Debug("sign upload rejected: invalid biz")
		return "", "", err
	}

	resp, err := svc.ossManager.Sign(dir, svc.uploadCallbackURL(biz))
	if err != nil {
		logger.Error("sign upload failed", "error", err, "dir", dir)
		return "", "", errs.ErrInternal
	}
	return resp, dir, nil
}

func (svc *ossService) UploadCallback(ctx context.Context, bucket string, objectName string, size string) (domain.CallbackResult, error) {
	logger := serviceLogger("UploadCallback").With("bucket", bucket, "object", objectName)
	if bucket == "" || bucket != svc.bucket || objectName == "" {
		logger.Debug("upload callback rejected: invalid bucket or object")
		return domain.CallbackResult{}, errs.ErrInvalidArgument
	}

	result, err := parseObject(objectName)
	if err != nil {
		logger.Debug("upload callback rejected: invalid object")
		return domain.CallbackResult{}, err
	}
	result.Size = size
	return result, nil
}

func (svc *ossService) GetObjectURL(ctx context.Context, objectName string) (string, error) {
	logger := serviceLogger("GetObjectURL").With("object", objectName)
	if _, err := parseObject(objectName); err != nil {
		logger.Debug("get object url rejected: invalid object")
		return "", errs.ErrInvalidArgument
	}

	url, err := svc.ossManager.Resign(objectName)
	if err != nil {
		logger.Error("resign object url failed", "error", err)
		return "", errs.ErrInternal
	}
	return url, nil
}

// uploadCallbackURL 根据上传业务选择 OSS 回调地址
func (svc *ossService) uploadCallbackURL(biz domain.UploadBiz) string {
	const (
		userCallbackPath     = "/api/v1/users/callback"
		questionCallbackPath = "/api/v1/interviews/questions/callback"
	)

	base := svc.callbackURL
	if base == "" {
		base = "http://gopostery.top" + userCallbackPath
	}

	switch biz {
	case domain.UploadBizInterviewQuestionBank:
		if strings.Contains(base, questionCallbackPath) {
			return base
		}
		if strings.Contains(base, userCallbackPath) {
			return strings.Replace(base, userCallbackPath, questionCallbackPath, 1)
		}
		return strings.TrimRight(base, "/") + questionCallbackPath
	default:
		return base
	}
}

func uploadDir(biz domain.UploadBiz, userID int64) (string, error) {
	switch biz {
	case domain.UploadBizUserAvatar:
		return "users/avatar/" + strconv.FormatInt(userID, 10) + "/", nil
	case domain.UploadBizInterviewQuestionBank:
		return "interviews/questions/" + strconv.FormatInt(userID, 10) + "/", nil
	default:
		return "", errs.ErrInvalidArgument
	}
}

func parseObject(objectName string) (domain.CallbackResult, error) {
	segments := strings.Split(objectName, "/")
	if len(segments) != 4 {
		return domain.CallbackResult{}, errs.ErrInvalidArgument
	}
	if segments[3] == "" {
		return domain.CallbackResult{}, errs.ErrInvalidArgument
	}

	userID, err := strconv.ParseInt(segments[2], 10, 64)
	if err != nil || userID <= 0 {
		return domain.CallbackResult{}, errs.ErrInvalidArgument
	}

	switch {
	case segments[0] == "users" && segments[1] == "avatar":
		return domain.CallbackResult{
			Biz:        domain.UploadBizUserAvatar,
			UserID:     userID,
			ObjectName: objectName,
			SourceFile: path.Base(objectName),
		}, nil
	case segments[0] == "interviews" && segments[1] == "questions":
		return domain.CallbackResult{
			Biz:        domain.UploadBizInterviewQuestionBank,
			UserID:     userID,
			ObjectName: objectName,
			SourceFile: path.Base(objectName),
		}, nil
	default:
		return domain.CallbackResult{}, errs.ErrInvalidArgument
	}
}
