package service

import (
	"context"
	"errors"
	"log/slog"

	dto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
)

var (
	ErrDuplicatedFollow    = errors.New("重复关注")
	ErrDuplicatedDisFollow = errors.New("重复取消关注")
)

type followService struct {
	FollowRepo repository.FollowRepository
	UserRepo   repository.UserRepository
	idGen      IDGenerator
}

func NewFollowService(followRepo repository.FollowRepository, userRepo repository.UserRepository, idGen IDGenerator) FollowService {
	return &followService{
		FollowRepo: followRepo,
		UserRepo:   userRepo,
		idGen:      idGen,
	}
}

// Follow 关注
func (svc *followService) Follow(ctx context.Context, ferId, feeId int64) error {
	res, err := svc.FollowRepo.Exists(ctx, ferId, feeId)
	if err != nil {
		return errno.ErrServerInternal // 数据库内部错误
	}

	if res == 1 || res == 3 { // 已经关注过了
		return errno.ErrDuplicatedFollow
	}

	follow := &model.Follow{
		ID:         svc.idGen.NextID(),
		FollowerID: ferId,
		FolloweeID: feeId,
	}
	err = svc.FollowRepo.Create(ctx, follow)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			// 检查过仍冲突
			slog.Error("Follow Failed", "error", err)
			return errno.ErrDuplicatedFollow
		}
		return errno.ErrServerInternal
	}

	return nil
}

// UnFollow 取消关注
func (svc *followService) UnFollow(ctx context.Context, ferId, feeId int64) error {
	res, err := svc.FollowRepo.Exists(ctx, ferId, feeId)
	if err != nil {
		return errno.ErrServerInternal // 数据库内部错误
	}

	if res == 2 || res == 0 { // 只有对方关注了我，或者互不关注
		return errno.ErrDuplicatedUnFollow
	}

	err = svc.FollowRepo.Delete(ctx, ferId, feeId)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Error("检查过还出错", "error", err)
			return errno.ErrDuplicatedUnFollow
		}
		return errno.ErrServerInternal
	}

	return nil
}

// IfFollow 判断关注关系
func (svc *followService) IfFollow(ctx context.Context, ferId, feeId int64) (model.FollowType, error) {
	res, err := svc.FollowRepo.Exists(ctx, ferId, feeId)
	if err != nil {
		return -1, errno.ErrServerInternal // 数据库内部错误
	}
	return res, nil
}

// GetFollowersByPage 按页查找粉丝
func (svc *followService) ListFollowersByPage(ctx context.Context, uid int64, pageNo, pageSize int) (int, []dto.BriefDTO, error) {
	var empty []dto.BriefDTO
	total, followersId, err := svc.FollowRepo.GetFollowers(ctx, uid, pageNo, pageSize)
	if err != nil {
		return 0, empty, errno.ErrServerInternal
	}

	res := make([]dto.BriefDTO, 0)
	for _, id := range followersId {
		user, err := svc.UserRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		userBriefDTO := dto.ToBriefDTO(user)
		res = append(res, userBriefDTO)
	}

	return int(total), res, nil
}

// GetFolloweesByPage 按页查找关注对象
func (svc *followService) ListFolloweesByPage(ctx context.Context, uid int64, pageNo, pageSize int) (int, []dto.BriefDTO, error) {
	var empty []dto.BriefDTO
	total, followeesId, err := svc.FollowRepo.GetFollowees(ctx, uid, pageNo, pageSize)
	if err != nil {
		return 0, empty, errno.ErrServerInternal
	}

	res := make([]dto.BriefDTO, 0)
	for _, id := range followeesId {
		user, err := svc.UserRepo.GetByID(ctx, id)
		if err != nil {

			continue
		}
		userBriefDTO := dto.ToBriefDTO(user)
		res = append(res, userBriefDTO)
	}

	return int(total), res, nil
}
