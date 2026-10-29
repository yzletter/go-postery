package repository

import (
	"context"
	"log/slog"
	"time"

	event_model "github.com/yzletter/go-postery/backend/event/outbox/model"
	"github.com/yzletter/go-postery/backend/micro/post/dto"
	"github.com/yzletter/go-postery/backend/micro/post/model"
	"github.com/yzletter/go-postery/backend/micro/post/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/post/repository/dao"
	"github.com/yzletter/go-postery/backend/ports"
	"github.com/yzletter/go-postery/backend/utils"
)

type postRepository struct {
	dao   dao.PostDAO
	cache cache.PostCache
	idGen ports.IDGenerator
}

func NewPostRepository(postDao dao.PostDAO, postCache cache.PostCache, idGen ports.IDGenerator) PostRepository {
	return &postRepository{dao: postDao, cache: postCache, idGen: idGen}
}

// Create 创建帖子
func (repo *postRepository) Create(ctx context.Context, post dto.Post, events []*event_model.OutboxEvent) error {
	// 构造 Post、Tag 和 PostTag
	m := dto.ToModelPost(post)
	tags := make([]*model.Tag, 0, len(post.Tags))
	postTags := make([]*model.PostTag, 0, len(post.Tags))
	for _, tagName := range post.Tags {
		tags = append(tags, &model.Tag{
			ID:   repo.idGen.NextID(),
			Name: tagName,
			Slug: utils.Slugify(tagName),
		})
		postTags = append(postTags, &model.PostTag{
			ID:     repo.idGen.NextID(),
			PostID: post.ID,
		})
	}

	// 写入 MySQL
	if err := repo.dao.Create(ctx, &m, tags, postTags, events); err != nil {
		return toRepositoryErr(err)
	}

	// 删除作者首页缓存
	if err := repo.cache.DelAuthorHomePage(ctx, post.User.UserID); err != nil {
		slog.Warn("delete author home page post cache failed", "author_id", post.User.UserID, "error", err)
	}

	return nil
}

// Delete 删除帖子
func (repo *postRepository) Delete(ctx context.Context, postID int64, authorID int64, events []*event_model.OutboxEvent) error {
	if err := repo.dao.Delete(ctx, postID, events); err != nil {
		return toRepositoryErr(err)
	}

	// 删除 Rank 分数
	if err := repo.cache.DeleteScore(ctx, postID); err != nil {
		slog.Warn("delete post rank score failed", "post_id", postID, "error", err)
	}

	// 删除作者首页缓存
	if err := repo.cache.DelAuthorHomePage(ctx, authorID); err != nil {
		slog.Warn("delete author home page post cache failed", "author_id", authorID, "error", err)
	}

	// 删帖子缓存
	if err := repo.cache.DelPost(ctx, postID); err != nil {
		slog.Warn("delete post cache failed", "post_id", postID, "error", err)
	}
	return nil
}

// Update 更新帖子
func (repo *postRepository) Update(ctx context.Context, post dto.Post, events []*event_model.OutboxEvent) error {
	// 构造 Post、Tag 和 PostTag
	m := dto.ToModelPost(post)
	tags := make([]*model.Tag, 0, len(post.Tags))
	postTags := make([]*model.PostTag, 0, len(post.Tags))
	for _, tagName := range post.Tags {
		tags = append(tags, &model.Tag{
			ID:   repo.idGen.NextID(),
			Name: tagName,
			Slug: utils.Slugify(tagName),
		})
		postTags = append(postTags, &model.PostTag{
			ID:     repo.idGen.NextID(),
			PostID: post.ID,
		})
	}

	// 更新帖子和标签
	if err := repo.dao.Update(ctx, &m, tags, postTags, events); err != nil {
		return toRepositoryErr(err)
	}

	// 删除作者首页缓存
	if err := repo.cache.DelAuthorHomePage(ctx, post.User.UserID); err != nil {
		slog.Warn("delete author home page post cache failed", "author_id", post.User.UserID, "error", err)
	}

	// 删帖子缓存
	if err := repo.cache.DelPost(ctx, post.ID); err != nil {
		slog.Warn("delete post cache failed", "post_id", post.ID, "error", err)
	}
	return nil
}

// GetByID 根据帖子 ID 获取帖子
func (repo *postRepository) GetByID(ctx context.Context, id int64) (dto.Post, error) {
	// 查缓存
	if post, err := repo.cache.GetPost(ctx, id); err == nil {
		return dto.ToPost(post), nil
	}

	// 查数据库
	post, err := repo.dao.GetByID(ctx, id)
	if err != nil {
		return dto.Post{}, toRepositoryErr(err)
	}

	// 写帖子缓存
	if err := repo.cache.SetPost(ctx, id, post); err != nil {
		slog.Warn("set post cache failed", "post_id", id, "error", err)
	}

	// 返回结果
	return dto.ToPost(post), nil
}

// GetPostByTime 根据时间查找帖子 ID
func (repo *postRepository) GetPostByTime(ctx context.Context, timeAt time.Time) ([]int64, error) {
	ids, err := repo.dao.GetPostByTime(ctx, timeAt)
	if err != nil {
		return nil, toRepositoryErr(err)
	}
	return ids, nil
}

// GetAuthorHomePage 获取用户首页帖子
func (repo *postRepository) GetAuthorHomePage(ctx context.Context, userID int64) ([]dto.Post, error) {
	// 查缓存
	if posts, err := repo.cache.GetAuthorHomePage(ctx, userID); err == nil {
		return dto.ToPosts(posts), nil
	}

	// 查数据库
	_, posts, err := repo.GetByAuthor(ctx, userID, 1, 10)
	if err != nil {
		return nil, err
	}

	// 更新作者首页缓存
	if err := repo.cache.SetAuthorHomePage(ctx, userID, dto.ToModelPosts(posts)); err != nil {
		slog.Warn("set author home page post cache failed", "author_id", userID, "error", err)
	}

	return posts, nil
}

// GetByAuthor 根据作者按页获取帖子
func (repo *postRepository) GetByAuthor(ctx context.Context, id int64, pageNo, pageSize int) (int64, []dto.Post, error) {
	total, posts, err := repo.dao.GetByUid(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	return total, dto.ToPosts(posts), nil
}

// GetByPage 按页获取帖子
func (repo *postRepository) GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []dto.Post, error) {
	total, posts, err := repo.dao.GetByPage(ctx, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	return total, dto.ToPosts(posts), nil
}

// GetByPageAndTag 按页 + Tag获取帖子
func (repo *postRepository) GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []dto.Post, error) {
	total, posts, err := repo.dao.GetByPageAndTag(ctx, tid, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	return total, dto.ToPosts(posts), nil
}
