package service

import (
	"errors"
	"log/slog"

	followRepository "github.com/yzletter/go-postery/repository/follow"
	userRepository "github.com/yzletter/go-postery/repository/user"
)

var (
	ErrDuplicatedFollow = errors.New("重复关注")
)

type FollowService struct {
	FollowDBRepo    *followRepository.FollowDBRepository
	FollowCacheRepo *followRepository.FollowCacheRepository
}

func NewFollowService(followDBRepo *followRepository.FollowDBRepository, followCacheRepo *followRepository.FollowCacheRepository) *FollowService {
	return &FollowService{FollowDBRepo: followDBRepo, FollowCacheRepo: followCacheRepo}
}

func (svc *FollowService) Follow(ferId, feeId int) error {
	res, err := svc.FollowDBRepo.IfFollow(ferId, feeId)
	if err != nil {
		return userRepository.ErrMySQLInternal // 数据库内部错误
	}

	if res == 1 || res == 3 { // 已经关注过了
		return ErrDuplicatedFollow
	}

	err = svc.FollowDBRepo.Follow(ferId, feeId)
	if err != nil {
		if errors.Is(err, userRepository.ErrUniqueKeyConflict) {
			slog.Error("检查过还出错", "error", err)
			return ErrDuplicatedFollow
		}
		return errors.New("关注失败")
	}

	return nil
}
func (svc *FollowService) IfFollow(ferId, feeId int) (int, error) {
	res, err := svc.FollowDBRepo.IfFollow(ferId, feeId)
	if err != nil {
		return 0, userRepository.ErrMySQLInternal // 数据库内部错误
	}

	return res, nil
}
