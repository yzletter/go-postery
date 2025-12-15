package service

import (
	"errors"
	"fmt"
	"log/slog"

	dto "github.com/yzletter/go-postery/dto/response"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/repository/dao"
	"gorm.io/gorm"
)

var (
	ErrDuplicatedFollow    = errors.New("重复关注")
	ErrDuplicatedDisFollow = errors.New("重复取消关注")
)

type followService struct {
	FollowRepo repository.FollowRepository
	UserRepo   repository.UserRepository
}

func NewFollowService(followRepo repository.FollowRepository, userRepo repository.UserRepository) FollowService {
	return &followService{
		FollowRepo: followRepo,
		UserRepo:   userRepo,
	}
}

func (svc *followService) Follow(ferId, feeId int) error {
	res, err := svc.FollowRepo.IfFollow(ferId, feeId)
	if err != nil {
		return dao.ErrInternal // 数据库内部错误
	}

	if res == 1 || res == 3 { // 已经关注过了
		return ErrDuplicatedFollow
	}

	err = svc.FollowRepo.Follow(ferId, feeId)
	if err != nil {
		if errors.Is(err, dao.ErrUniqueKeyConflict) {
			slog.Error("检查过还出错", "error", err)
			return ErrDuplicatedFollow
		}
		return errors.New("关注失败, 请稍后重试")
	}

	return nil
}

func (svc *followService) DisFollow(ferId, feeId int) error {
	res, err := svc.FollowRepo.IfFollow(ferId, feeId)
	if err != nil {
		return dao.ErrInternal // 数据库内部错误
	}

	if res == 2 || res == 0 { // 只有对方关注了我，或者互不关注
		return ErrDuplicatedDisFollow
	}

	err = svc.FollowRepo.DisFollow(ferId, feeId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("检查过还出错", "error", err)
			return ErrDuplicatedDisFollow
		}
		return errors.New("取消关注失败")
	}

	return nil
}

func (svc *followService) IfFollow(ferId, feeId int) (int, error) {
	res, err := svc.FollowRepo.IfFollow(ferId, feeId)
	if err != nil {
		return 0, dao.ErrInternal // 数据库内部错误
	}

	return res, nil
}

func (svc *followService) GetFollowers(uid int) ([]dto.UserBriefDTO, error) {
	followersId, err := svc.FollowRepo.GetFollowers(uid)
	fmt.Println(followersId)
	if err != nil {
		return nil, dao.ErrInternal
	}

	res := make([]dto.UserBriefDTO, 0)
	for _, id := range followersId {
		user, err := svc.UserRepo.GetByID(int64(id))
		if err != nil {
			continue
		}
		userBriefDTO := dto.ToUserBriefDTO(*user)
		res = append(res, userBriefDTO)
	}
	fmt.Println(res)

	return res, nil
}

func (svc *followService) GetFollowees(uid int) ([]dto.UserBriefDTO, error) {
	followeesId, err := svc.FollowRepo.GetFollowees(uid)
	if err != nil {
		return nil, dao.ErrInternal
	}

	res := make([]dto.UserBriefDTO, 0)
	for _, id := range followeesId {
		user, err := svc.UserRepo.GetByID(int64(id))
		if err != nil {

			continue
		}
		userBriefDTO := dto.ToUserBriefDTO(*user)
		res = append(res, userBriefDTO)
	}

	return res, nil
}
