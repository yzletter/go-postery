package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
)

const OneWeekTimeSecs = 60 * 60 * 24 * 7

type postRepository struct {
	dao   dao.PostDAO
	cache cache.PostCache
}

func NewPostRepository(postDao dao.PostDAO, postCache cache.PostCache) PostRepository {
	return &postRepository{dao: postDao, cache: postCache}
}

func (repo *postRepository) Create(ctx context.Context, post *model.Post) error {
	// 创建文章
	err := repo.dao.Create(ctx, post)
	if err != nil {
		return toRepositoryErr(err)
	}

	// todo 写 Cache

	// 初始化文章分数
	err = repo.cache.SetScore(ctx, post.ID)
	if err != nil {
		return ErrServerInternal
	}
	return nil
}

func (repo *postRepository) Delete(ctx context.Context, id int64) error {
	err := repo.dao.Delete(ctx, id)
	if err != nil {
		return toRepositoryErr(err)
	}

	// todo 删 Cache

	return nil
}

func (repo *postRepository) UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error {
	// DAO
	err := repo.dao.UpdateCount(ctx, id, field, delta)
	if err != nil {
		return toRepositoryErr(err)
	}

	post, _ := repo.dao.GetByID(ctx, id)

	// Cache
	ok, err := repo.cache.ChangeInteractiveCnt(ctx, id, field, delta)
	if err != nil || !ok {
		// 获取失败的 Field 名
		col, colErr := field.Column()
		if colErr != nil {
			col = "invalid"
		}
		slog.Error("Cache ChangeInteractiveCnt Failed", "id", id, "field", col, "delta", delta, "error", err)
		if post != nil {
			fields := []model.PostCntField{model.PostViewCount, model.PostCommentCount, model.PostLikeCount}
			vals := []int{post.ViewCount, post.ViewCount, post.LikeCount}
			repo.cache.SetInteractiveKey(ctx, id, fields, vals)
		}
	}

	return nil
}

func (repo *postRepository) Update(ctx context.Context, id int64, updates map[string]any) error {
	err := repo.dao.Update(ctx, id, updates)
	if err != nil {
		return toRepositoryErr(err)
	}

	// todo 改Cache

	return nil
}

func (repo *postRepository) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	// todo 读 Cache

	post, err := repo.dao.GetByID(ctx, id)
	if err != nil {
		return nil, toRepositoryErr(err)
	}

	return post, nil
}

func (repo *postRepository) GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error) {
	// todo 读 Cache

	total, posts, err := repo.dao.GetByUid(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	return total, posts, nil
}

func (repo *postRepository) GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error) {
	// todo 读 Cache

	total, posts, err := repo.dao.GetByPage(ctx, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	return total, posts, nil
}

func (repo *postRepository) GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error) {
	// todo 读 Cache

	total, posts, err := repo.dao.GetByPageAndTag(ctx, tid, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	return total, posts, nil
}

// ChangeScore 修改帖子分数
func (repo *postRepository) ChangeScore(ctx context.Context, pid int64, delta int) {
	// 查询是否在热度期内
	value, err := repo.cache.CheckPostLikeTime(ctx, pid)
	if err != nil {
		slog.Error("Check Post Like Time Failed", "error", err)
		return
	}

	// 过了热度期，热度不再波动
	if float64(time.Now().Unix())-value > OneWeekTimeSecs {
		return
	}

	err = repo.cache.ChangeScore(ctx, pid, delta)
	if err != nil {
		slog.Error("Change Post Score Failed", "error", err)
		return
	}
}

func (repo *postRepository) Top(ctx context.Context) ([]*model.Post, []float64, error) {
	ids, scores, err := repo.cache.Top(ctx)
	if err != nil {
		return nil, nil, ErrServerInternal
	}

	var posts []*model.Post
	for _, id := range ids {
		post, err := repo.dao.GetByID(ctx, id)
		if err != nil {
			post = &model.Post{
				ID:    0,
				Title: "未知文章",
			}
		}
		posts = append(posts, post)
	}

	return posts, scores, nil
}
