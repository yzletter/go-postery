package repository

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/rank/domain"
	"github.com/yzletter/go-postery/backend/micro/rank/repository/cache"
)

type rankRepository struct {
	cache cache.RankCache
}

func NewRankRepository(cache cache.RankCache) RankRepository {
	return &rankRepository{
		cache: cache,
	}
}

func (repo *rankRepository) UpdateUserScore(ctx context.Context, id int64, score int64) error {
	err := repo.cache.UpdateScore(ctx, domain.BizUser, id, score)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *rankRepository) UpdatePostScore(ctx context.Context, id int64, score int64) error {
	err := repo.cache.UpdateScore(ctx, domain.BizPost, id, score)
	if err != nil {
		return toRepositoryErr(err)
	}
	return nil
}

func (repo *rankRepository) GetTopKUser(ctx context.Context) ([]domain.User, error) {
	ids, scores, err := repo.cache.TopK(ctx, domain.BizUser, 10)
	if err != nil {
		return nil, toRepositoryErr(err)
	}

	res := make([]domain.User, 0, len(ids))
	for i := range len(ids) {
		res = append(res, domain.User{
			ID:    ids[i],
			Score: scores[i],
		})
	}
	return res, nil
}

func (repo *rankRepository) GetTopKPost(ctx context.Context) ([]domain.Post, error) {
	ids, scores, err := repo.cache.TopK(ctx, domain.BizPost, 10)
	if err != nil {
		return nil, toRepositoryErr(err)
	}

	res := make([]domain.Post, 0, len(ids))
	for i := range len(ids) {
		res = append(res, domain.Post{ID: ids[i], Score: scores[i]})
	}
	return res, nil
}
