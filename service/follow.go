package service

import repository "github.com/yzletter/go-postery/repository/follow"

type FollowService struct {
	FollowDBRepo    *repository.FollowDBRepository
	FollowCacheRepo *repository.FollowCacheRepository
}

func NewFollowService(followDBRepo *repository.FollowDBRepository, followCacheRepo *repository.FollowCacheRepository) *FollowService {
	return &FollowService{FollowDBRepo: followDBRepo, FollowCacheRepo: followCacheRepo}
}
