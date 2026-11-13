package repository

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/lottery/model"
	"github.com/yzletter/go-postery/backend/micro/lottery/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/lottery/repository/dao"
)

type giftRepository struct {
	dao   dao.GiftDAO
	cache cache.GiftCache
}

func NewGiftRepository(dao dao.GiftDAO, cache cache.GiftCache) GiftRepository {
	return &giftRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *giftRepository) Lottery(ctx context.Context, requestID int64, userID int64, giftID int64) (int64, int64, error) {
	requestID, giftID, err := repo.cache.Lottery(ctx, requestID, userID, giftID)
	if err != nil {
		return 0, 0, toRepositoryErr(err)
	}
	return requestID, giftID, nil
}

func (repo *giftRepository) Reuse(ctx context.Context, requestID int64, userID int64) (int64, int64, error) {
	requestID, giftID, err := repo.cache.Reuse(ctx, requestID, userID)
	if err != nil {
		return 0, 0, toRepositoryErr(err)
	}
	return requestID, giftID, nil
}

func (repo *giftRepository) ScanLotteryRecords(ctx context.Context, cursor uint64, count int64) ([]*model.LotteryRecord, uint64, error) {
	records, nextCursor, err := repo.cache.ScanLotteryRecords(ctx, cursor, count)
	if err != nil {
		return nil, cursor, toRepositoryErr(err)
	}
	return records, nextCursor, nil
}

func (repo *giftRepository) GetAllGifts(ctx context.Context) ([]*model.Gift, error) {
	gifts, err := repo.dao.GetAll(ctx)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return gifts, nil
}

func (repo *giftRepository) GetCacheInventory(ctx context.Context) ([]*model.Gift, error) {
	gifts, err := repo.cache.GetAllInventory(ctx)
	if err != nil {
		return nil, ErrRecordNotFound
	}
	return gifts, nil
}

func (repo *giftRepository) GetByID(ctx context.Context, gid int64) (*model.Gift, error) {
	gift, err := repo.dao.GetByID(ctx, gid)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return gift, nil
}

func (repo *giftRepository) RollbackCacheInventory(ctx context.Context, orderID, gid int64) error {
	err := repo.cache.RollbackInventory(ctx, orderID, gid)
	if err != nil {
		return ErrResourceConflict
	}
	return nil
}

func (repo *giftRepository) InitCacheInventory(ctx context.Context) {
	gifts, err := repo.dao.GetAll(ctx)
	if err != nil {
		return
	}
	repo.cache.InitInventory(ctx, gifts)
}
