package repository

import (
	"context"
	"log/slog"

	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
)

type postRepository struct {
	dao   dao.PostDAO
	cache cache.PostCache
}

func NewPostRepository(postDao dao.PostDAO, postCache cache.PostCache) PostRepository {
	return &postRepository{dao: postDao, cache: postCache}
}

func (repo *postRepository) Create(ctx context.Context, post *model.Post) (*model.Post, error) {
	p, err := repo.dao.Create(ctx, post)
	if err != nil {
		return nil, toRepoErr(err)
	}

	// todo 写 Cache

	return p, nil
}

func (repo *postRepository) Delete(ctx context.Context, id int64) error {
	err := repo.dao.Delete(ctx, id)
	if err != nil {
		return toRepoErr(err)
	}

	// todo 删 Cache

	return nil
}

func (repo *postRepository) UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error {
	// DAO
	err := repo.dao.UpdateCount(ctx, id, field, delta)
	if err != nil {
		return toRepoErr(err)
	}

	// Cache
	ok, err := repo.cache.ChangeInteractiveCnt(ctx, id, field, delta)
	if err != nil || !ok {
		// 获取失败的 Field 名
		col, colErr := field.Column()
		if colErr != nil {
			col = "invalid"
		}
		slog.Error("Cache ChangeInteractiveCnt Failed", "id", id, "field", col, "delta", delta, "error", err)
	}

	return nil
}

func (repo *postRepository) Update(ctx context.Context, id int64, updates map[string]any) error {
	err := repo.dao.Update(ctx, id, updates)
	if err != nil {
		return toRepoErr(err)
	}

	// todo 改Cache

	return nil
}

func (repo *postRepository) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	// todo 读 Cache

	post, err := repo.dao.GetByID(ctx, id)
	if err != nil {
		return nil, toRepoErr(err)
	}

	return post, nil
}

func (repo *postRepository) GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error) {
	// todo 读 Cache

	total, posts, err := repo.dao.GetByUid(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepoErr(err)
	}

	return total, posts, nil
}

func (repo *postRepository) GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error) {
	// todo 读 Cache

	total, posts, err := repo.dao.GetByPage(ctx, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepoErr(err)
	}

	return total, posts, nil
}

func (repo *postRepository) GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error) {
	// todo 读 Cache

	total, posts, err := repo.dao.GetByPageAndTag(ctx, tid, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepoErr(err)
	}

	return total, posts, nil
}
