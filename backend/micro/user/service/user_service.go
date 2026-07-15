package service

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	interactive_grpc "github.com/yzletter/go-postery/api/proto/interactive/v1"
	oss_grpc "github.com/yzletter/go-postery/api/proto/oss/v1"
	rank_grpc "github.com/yzletter/go-postery/api/proto/rank/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/user/domain"
	"github.com/yzletter/go-postery/backend/micro/user/repository"
	"github.com/yzletter/go-postery/backend/ports"
)

type userService struct {
	userRepo    repository.UserRepository
	interClient manager.InteractiveClient
	rankClient  manager.RankClient
	ossClient   manager.OSSClient
	idGen       ports.IDGenerator
}

// NewUserService 构造函数
func NewUserService(userRepository repository.UserRepository, interClient manager.InteractiveClient, rankClient manager.RankClient, ossClient manager.OSSClient, idGen ports.IDGenerator) UserService {
	return &userService{
		userRepo:    userRepository,
		interClient: interClient,
		rankClient:  rankClient,
		ossClient:   ossClient,
		idGen:       idGen,
	}
}

// serviceLogger 构造用户服务日志
func serviceLogger(method string) *slog.Logger {
	return slog.With("component", "user_service", "method", method)
}

// GetProfile 获取用户资料
func (svc *userService) GetProfile(ctx context.Context, id int64) (domain.Profile, error) {
	logger := serviceLogger("GetProfile").With("user_id", id)
	if id <= 0 {
		logger.Debug("get profile rejected: invalid user id")
		return domain.Profile{}, errs.ErrInvalidArgument
	}

	// 获取用户资料
	profile, err := svc.userRepo.GetProfile(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			logger.Debug("profile not found")
			return domain.Profile{}, errs.ErrNotFound
		}
		logger.Error("get profile failed", "error", err)
		return domain.Profile{}, errs.ErrInternal
	}

	return profile, nil
}

// UploadAvatarSign 获取上传头像 OSS 签名
func (svc *userService) UploadAvatarSign(ctx context.Context, id int64) (string, error) {
	logger := serviceLogger("UploadAvatarSign").With("user_id", id)
	if id <= 0 {
		logger.Debug("sign avatar upload rejected: invalid user id")
		return "", errs.ErrInvalidArgument
	}

	resp, err := svc.ossClient.SignUpload(ctx, &oss_grpc.SignUploadRequest{
		Biz:      1,
		UserID:   id,
		FileName: "",
	})
	if err != nil {
		logger.Error("sign avatar upload failed", "error", err)
		return "", errs.ErrInternal
	}
	return resp.Response, err
}

// UploadAvatarCallback 处理头像上传回调
func (svc *userService) UploadAvatarCallback(ctx context.Context, id int64, object string) error {
	logger := serviceLogger("UploadAvatarCallback").With("user_id", id, "object", object)
	if id <= 0 {
		logger.Debug("avatar callback rejected: invalid user id")
		return errs.ErrInvalidArgument
	}

	// 校验前缀
	prefix := "users/avatar/" + strconv.FormatInt(id, 10) + "/"
	if object == "" || !strings.HasPrefix(object, prefix) {
		logger.Debug("avatar callback rejected: invalid object")
		return errs.ErrInvalidArgument
	}

	// 更新头像对象名
	if err := svc.userRepo.UpdateProfile(ctx, id, map[string]any{"avatar": object}); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			logger.Debug("avatar callback rejected: profile not found")
			return errs.ErrNotFound
		}
		logger.Error("update avatar failed", "error", err)
		return errs.ErrInternal
	}
	return nil
}

// GetAvatarURL 获取头像访问预签名 URL
func (svc *userService) GetAvatarURL(ctx context.Context, object string) (string, error) {
	logger := serviceLogger("GetAvatarURL").With("object", object)
	if object == "" || !strings.HasPrefix(object, "users/avatar/") {
		logger.Debug("get avatar url rejected: invalid object")
		return "", errs.ErrInvalidArgument
	}

	resp, err := svc.ossClient.GetObjectURL(ctx, &oss_grpc.GetObjectURLRequest{ObjectName: object})
	if err != nil {
		logger.Error("resign avatar url failed", "error", err)
		return "", errs.ErrInternal
	}
	return resp.URL, nil
}

// GetIDAfterTime 根据时间获取之后创建的用户 ID
func (svc *userService) GetIDAfterTime(ctx context.Context, timeAt time.Time) ([]int64, error) {
	logger := serviceLogger("GetIDAfterTime").With("time_after", timeAt)
	if timeAt.IsZero() {
		logger.Debug("get user ids rejected: invalid time")
		return nil, errs.ErrInvalidArgument
	}

	ids, err := svc.userRepo.GetIDAfterTime(ctx, timeAt)
	if err != nil {
		logger.Error("get user ids after time failed", "error", err)
		return nil, errs.ErrInternal
	}
	return ids, nil
}

// UpdateProfile 更新用户资料
func (svc *userService) UpdateProfile(ctx context.Context, id int64, updates map[string]any) error {
	logger := serviceLogger("UpdateProfile").With("user_id", id)
	if id <= 0 {
		logger.Debug("update profile rejected: invalid user id")
		return errs.ErrInvalidArgument
	}

	if len(updates) == 0 {
		return nil
	}

	if err := svc.userRepo.UpdateProfile(ctx, id, updates); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			logger.Debug("update profile rejected: profile not found")
			return errs.ErrNotFound
		}
		if errors.Is(err, repository.ErrUniqueKey) {
			logger.Debug("update profile rejected: unique key conflict")
			return errs.ErrAlreadyExits
		}
		logger.Error("update profile failed", "error", err)
		return errs.ErrInternal
	}

	return nil
}

// ListFollowers 按页查找粉丝
func (svc *userService) ListFollowers(ctx context.Context, id int64, pageNo int, pageSize int) (int64, []domain.ProfileBrief, error) {
	logger := serviceLogger("ListFollowers").With("user_id", id, "page_no", pageNo, "page_size", pageSize)
	if id <= 0 || pageNo < 1 || pageSize < 1 || pageSize > 100 {
		logger.Debug("list followers rejected: invalid params")
		return 0, nil, errs.ErrInvalidArgument
	}

	resp, err := svc.interClient.GetFollowers(ctx, &interactive_grpc.ListFollowRequest{UserID: id, PageNo: uint32(pageNo), PageSize: uint32(pageSize)})
	if err != nil {
		logger.Error("get followers failed", "error", err)
		return 0, nil, errs.ErrInternal
	}

	// 获取粉丝资料，资料缺失时降级为未知用户
	profiles := make([]domain.ProfileBrief, 0, len(resp.IDs))
	fallback := 0
	for _, fid := range resp.IDs {
		if profile, err := svc.userRepo.GetProfile(ctx, fid); err == nil {
			profiles = append(profiles, profile.Briefed())
		} else {
			if errors.Is(err, repository.ErrRecordNotFound) {
				fallback++
				profiles = append(profiles, domain.ProfileBrief{UserID: fid, Nickname: "未知用户"})
				logger.Warn("use fallback follower profile", "follower_id", fid)
				continue
			}
			logger.Error("get follower profile failed", "follower_id", fid, "error", err)
			return int64(resp.Count), nil, errs.ErrInternal
		}
	}
	if fallback > 0 {
		logger.Warn("some follower profiles used fallback", "fallback", fallback, "requested", len(resp.IDs))
	}

	return int64(resp.Count), profiles, nil
}

// ListFollowees 按页查找关注的人
func (svc *userService) ListFollowees(ctx context.Context, id int64, pageNo int, pageSize int) (int64, []domain.ProfileBrief, error) {
	logger := serviceLogger("ListFollowees").With("user_id", id, "page_no", pageNo, "page_size", pageSize)
	if id <= 0 || pageNo < 1 || pageSize < 1 || pageSize > 100 {
		logger.Debug("list followees rejected: invalid params")
		return 0, nil, errs.ErrInvalidArgument
	}

	resp, err := svc.interClient.GetFollowees(ctx, &interactive_grpc.ListFollowRequest{UserID: id, PageNo: uint32(pageNo), PageSize: uint32(pageSize)})
	if err != nil {
		logger.Error("get followees failed", "error", err)
		return 0, nil, errs.ErrInternal
	}

	// 获取关注用户资料，资料缺失时降级为未知用户
	profiles := make([]domain.ProfileBrief, 0, len(resp.IDs))
	fallback := 0
	for _, fid := range resp.IDs {
		if profile, err := svc.userRepo.GetProfile(ctx, fid); err == nil {
			profiles = append(profiles, profile.Briefed())
		} else {
			if errors.Is(err, repository.ErrRecordNotFound) {
				fallback++
				profiles = append(profiles, domain.ProfileBrief{UserID: fid, Nickname: "未知用户"})
				logger.Warn("use fallback followee profile", "followee_id", fid)
				continue
			}
			logger.Error("get followee profile failed", "followee_id", fid, "error", err)
			return int64(resp.Count), nil, errs.ErrInternal
		}
	}
	if fallback > 0 {
		logger.Warn("some followee profiles used fallback", "fallback", fallback, "requested", len(resp.IDs))
	}

	return int64(resp.Count), profiles, nil
}

// Top 获取用户排行榜
func (svc *userService) Top(ctx context.Context) ([]domain.ProfileTop, error) {
	logger := serviceLogger("Top")

	resp, err := svc.rankClient.TopKUser(ctx, &rank_grpc.RankEmptyRequest{})
	if err != nil {
		logger.Error("get top users failed", "error", err)
		return nil, errs.ErrInternal
	}

	// 获取排行榜用户资料，资料缺失时降级为未知用户
	profiles := make([]domain.ProfileTop, 0, len(resp.Users))
	fallback := 0
	for _, user := range resp.Users {
		if profile, err := svc.userRepo.GetProfile(ctx, user.ID); err == nil {
			profiles = append(profiles, profile.Topped(user.Score))
		} else {
			if errors.Is(err, repository.ErrRecordNotFound) {
				fallback++
				profiles = append(profiles, domain.ProfileTop{
					ProfileBrief: domain.ProfileBrief{UserID: user.ID, Nickname: "未知用户"},
					Score:        user.Score,
				})
				logger.Warn("use fallback ranked user profile", "rank_user_id", user.ID, "score", user.Score)
				continue
			}
			logger.Error("get ranked user profile failed", "rank_user_id", user.ID, "score", user.Score, "error", err)
			return nil, errs.ErrInternal
		}
	}
	if fallback > 0 {
		logger.Warn("some ranked user profiles used fallback", "fallback", fallback, "requested", len(resp.Users))
	}
	return profiles, nil
}
