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

type FollowService struct {
	UserRepo   repository.UserRepository
	FollowRepo repository.FollowRepository
}

func NewFollowService(userRepo repository.UserRepository, followRepo repository.FollowRepository) *FollowService {
	return &FollowService{UserRepo: userRepo, FollowRepo: followRepo}
}

func (svc *FollowService) Follow(ferId, feeId int) error {
	res, err := svc.FollowDBRepo.IfFollow(ferId, feeId)
	if err != nil {
		return dao.ErrInternal // 数据库内部错误
	}

	if res == 1 || res == 3 { // 已经关注过了
		return ErrDuplicatedFollow
	}

	err = svc.FollowDBRepo.Follow(ferId, feeId)
	if err != nil {
		if errors.Is(err, dao.ErrUniqueKeyConflict) {
			slog.Error("检查过还出错", "error", err)
			return ErrDuplicatedFollow
		}
		return errors.New("关注失败, 请稍后重试")
	}

	return nil
}

func (svc *FollowService) DisFollow(ferId, feeId int) error {
	res, err := svc.FollowDBRepo.IfFollow(ferId, feeId)
	if err != nil {
		return dao.ErrInternal // 数据库内部错误
	}

	if res == 2 || res == 0 { // 只有对方关注了我，或者互不关注
		return ErrDuplicatedDisFollow
	}

	err = svc.FollowDBRepo.DisFollow(ferId, feeId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("检查过还出错", "error", err)
			return ErrDuplicatedDisFollow
		}
		return errors.New("取消关注失败")
	}

	return nil
}

func (svc *FollowService) IfFollow(ferId, feeId int) (int, error) {
	res, err := svc.FollowDBRepo.IfFollow(ferId, feeId)
	if err != nil {
		return 0, dao.ErrInternal // 数据库内部错误
	}

	return res, nil
}

func (svc *FollowService) GetFollowers(uid int) ([]dto.UserBriefDTO, error) {
	followersId, err := svc.FollowDBRepo.GetFollowers(uid)
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

func (svc *FollowService) GetFollowees(uid int) ([]dto.UserBriefDTO, error) {
	followeesId, err := svc.FollowDBRepo.GetFollowees(uid)
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
