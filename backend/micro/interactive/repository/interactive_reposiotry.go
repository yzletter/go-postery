package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
	"github.com/yzletter/go-postery/backend/micro/interactive/model"
	"github.com/yzletter/go-postery/backend/micro/interactive/repository/cache"
	"github.com/yzletter/go-postery/backend/micro/interactive/repository/dao"
)

type interactiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
}

func NewInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache) InteractiveRepository {
	return &interactiveRepository{
		dao:   dao,
		cache: cache,
	}
}

// GetPostInteractive 获取帖子互动信息
func (repo *interactiveRepository) GetPostInteractive(ctx context.Context, id int64) (domain.PostInter, error) {
	// 查缓存
	cachedInter, err := repo.cache.GetPostInteractive(ctx, id)
	if err == nil {
		return cachedInter, nil
	}

	// 查数据库
	inter, err := repo.dao.GetPostInteractive(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrRecordNotFound) {
			res := domain.PostInter{}
			if err := repo.cache.SetPostInteractive(ctx, id, res); err != nil {
				slog.Warn("set post interactive cache failed", "id", id, "error", err)
			}
			return res, nil
		}
		return domain.PostInter{}, toRepositoryErr(err)
	}
	res := domain.ToPostInterDomain(inter)

	// 更新缓存
	if err := repo.cache.SetPostInteractive(ctx, id, res); err != nil {
		slog.Warn("set post interactive cache failed", "id", id, "error", err)
	}

	return res, nil
}

// GetUserInteractive 获取用户互动信息
func (repo *interactiveRepository) GetUserInteractive(ctx context.Context, id int64) (domain.UserInter, error) {
	// 查缓存
	cachedInter, err := repo.cache.GetUserInteractive(ctx, id)
	if err == nil {
		return cachedInter, nil
	}

	// 查数据库
	inter, err := repo.dao.GetUserInteractive(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrRecordNotFound) {
			res := domain.UserInter{}
			if err := repo.cache.SetUserInteractive(ctx, id, res); err != nil {
				slog.Warn("set user interactive cache failed", "id", id, "error", err)
			}
			return res, nil
		}
		return domain.UserInter{}, toRepositoryErr(err)
	}
	res := domain.ToUserInterDomain(inter)

	// 更新缓存
	if err := repo.cache.SetUserInteractive(ctx, id, res); err != nil {
		slog.Warn("set user interactive cache failed", "id", id, "error", err)
	}

	return res, nil
}

// IncrReadCnt 增加帖子阅读数
func (repo *interactiveRepository) IncrReadCnt(ctx context.Context, consumer string, topic string, readEventPayloads ...*event.NewReadEventPayload) error {
	// 写数据库
	err := repo.dao.IncrReadCnt(ctx, consumer, topic, readEventPayloads...)
	if err != nil {
		return toRepositoryErr(err)
	}

	deleted := make(map[int64]struct{}, len(readEventPayloads))

	for _, payload := range readEventPayloads {
		// 判断是否删过缓存
		if _, ok := deleted[payload.PostID]; ok {
			continue
		}

		// 放入集合
		deleted[payload.PostID] = struct{}{}

		// 删缓存
		err = repo.cache.DelPostInteractive(ctx, payload.PostID)
		if err != nil {
			slog.Warn("delete post interactive cache failed", "id", payload.PostID, "error", err)
		}
	}
	// 删缓存

	return nil
}

func (repo *interactiveRepository) ChangeInteractiveCntWithOutbox(ctx context.Context, biz domain.BizType, bizID int64, timeAt time.Time, delta int64, processedEvent *event.ProcessedEvent) error {
	// 查缓存是否消费过
	ok, err := repo.cache.GetConsume(ctx, processedEvent.Consumer, processedEvent.EventID)
	if err == nil && ok {
		return nil
	}

	// 关注事件更新用户 Inter
	if biz == domain.BizFollow {
		// 获取用户 Inter 信息
		inter, err := repo.GetUserInteractive(ctx, bizID)
		if err != nil && !errors.Is(err, ErrRecordNotFound) {
			return err
		}
		// 已经被扫表统一计算过的消息直接丢弃
		if err == nil && timeAt.Before(inter.CalculateAt) {
			return nil
		}

		// 写数据库
		if err := repo.dao.ChangeInteractiveCnt(ctx, biz, bizID, delta, processedEvent); err != nil {
			return toRepositoryErr(err)
		}

		// 删缓存
		if err := repo.cache.DelUserInteractive(ctx, bizID); err != nil {
			slog.Warn("delete user interactive cache failed", "id", bizID, "error", err)
		}

		// 写缓存消费过
		if err := repo.cache.SetConsume(ctx, processedEvent.Consumer, processedEvent.EventID); err != nil {
			slog.Warn("mark consumed failed", "error", err, "event_id", processedEvent.EventID)
		}
		return nil
	}

	// 其它事件更新帖子 Inter
	inter, err := repo.GetPostInteractive(ctx, bizID)
	if err != nil && !errors.Is(err, ErrRecordNotFound) {
		return err
	}
	// 已经被扫表统一计算过的消息直接丢弃
	if err == nil && timeAt.Before(inter.CalculateAt) {
		return nil
	}

	// 写数据库
	if err := repo.dao.ChangeInteractiveCnt(ctx, biz, bizID, delta, processedEvent); err != nil {
		return toRepositoryErr(err)
	}

	// 删缓存
	if err := repo.cache.DelPostInteractive(ctx, bizID); err != nil {
		slog.Warn("delete post interactive cache failed", "id", bizID, "error", err)
	}

	// 写缓存消费过
	if err := repo.cache.SetConsume(ctx, processedEvent.Consumer, processedEvent.EventID); err != nil {
		slog.Warn("mark consumed failed", "error", err, "event_id", processedEvent.EventID)
	}
	return nil
}

// Like 点赞
func (repo *interactiveRepository) Like(ctx context.Context, like domain.Like, events ...*event.OutboxEvent) error {
	// 写数据库
	m := &model.Like{
		ID:     like.ID,
		UserID: like.UserID,
		PostID: like.PostID,
	}
	err := repo.dao.CreateLike(ctx, m, events...)
	if err != nil {
		return toRepositoryErr(err)
	}

	// 删缓存
	err = repo.cache.DelLike(ctx, like.UserID, like.PostID)
	if err != nil {
		slog.Warn("delete like cache failed", "user_id", like.UserID, "post_id", like.PostID, "error", err)
	}
	return nil
}

// UnLike 取消点赞
func (repo *interactiveRepository) UnLike(ctx context.Context, uid, pid int64, events ...*event.OutboxEvent) error {
	// 写数据库
	err := repo.dao.DelLike(ctx, uid, pid, events...)
	if err != nil {
		return toRepositoryErr(err)
	}

	// 删缓存
	err = repo.cache.DelLike(ctx, uid, pid)
	if err != nil {
		slog.Warn("delete like cache failed", "user_id", uid, "post_id", pid, "error", err)
	}
	return nil
}

// HasLiked 用户是否已点赞
func (repo *interactiveRepository) HasLiked(ctx context.Context, uid, pid int64) (bool, error) {
	// 查缓存
	cached, err := repo.cache.GetLike(ctx, uid, pid)
	if err == nil {
		return cached, nil
	}

	// 查数据库
	ok, err := repo.dao.GetLike(ctx, uid, pid)
	if err != nil {
		return false, toRepositoryErr(err)
	}

	// 更新缓存
	if ok {
		err = repo.cache.SetLike(ctx, uid, pid)
		if err != nil {
			slog.Warn("set like cache failed", "user_id", uid, "post_id", pid, "error", err)
		}
	}

	return ok, nil
}

// CreateFollow 新建关注关系
func (repo *interactiveRepository) CreateFollow(ctx context.Context, follow domain.Follow, events ...*event.OutboxEvent) error {
	// 写数据库
	m := &model.Follow{
		ID:         follow.ID,
		FollowerID: follow.FollowerID,
		FolloweeID: follow.FolloweeID,
	}
	err := repo.dao.CreateFollow(ctx, m, events...)
	if err != nil {
		return toRepositoryErr(err)
	}

	// 删缓存
	err = repo.cache.DelFollow(ctx, follow.FollowerID, follow.FolloweeID)
	if err != nil {
		slog.Warn("delete follow cache failed", "follower_id", follow.FollowerID, "followee_id", follow.FolloweeID, "error", err)
	}
	return nil
}

// DelFollow 删除关注关系
func (repo *interactiveRepository) DelFollow(ctx context.Context, ferID, feeID int64, events ...*event.OutboxEvent) error {
	// 写数据库
	err := repo.dao.DelFollow(ctx, ferID, feeID, events...)
	if err != nil {
		return toRepositoryErr(err)
	}

	// 删缓存
	err = repo.cache.DelFollow(ctx, ferID, feeID)
	if err != nil {
		slog.Warn("delete follow cache failed", "follower_id", ferID, "followee_id", feeID, "error", err)
	}
	return nil
}

// GetFollow 获取关注关系
func (repo *interactiveRepository) GetFollow(ctx context.Context, follower, followee int64) (domain.FollowType, error) {
	// 查缓存
	cachedType, err := repo.cache.GetFollow(ctx, follower, followee)
	if err == nil {
		return cachedType, nil
	}

	// 查数据库
	res, err := repo.dao.GetFollow(ctx, follower, followee)
	if err != nil {
		return -1, toRepositoryErr(err)
	}

	// 更新缓存
	if res != domain.FollowNone {
		var err error
		if res == domain.FollowIFollow {
			err = repo.cache.SetFollow(ctx, follower, followee)
		} else if res == domain.FollowFollowMe {
			err = repo.cache.SetFollow(ctx, followee, follower)
		} else {
			err = repo.cache.SetFollow(ctx, follower, followee)
			err = repo.cache.SetFollow(ctx, followee, follower)
		}
		if err != nil {
			slog.Warn("set follow cache failed", "follower_id", follower, "followee_id", followee, "error", err)
		}
	}

	return res, nil
}

// GetFollowers 按页获取用户关注的人
func (repo *interactiveRepository) GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	total, ids, err := repo.dao.GetFollowers(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	return total, ids, nil
}

// GetFollowees 按页获取用户粉丝
func (repo *interactiveRepository) GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error) {
	total, ids, err := repo.dao.GetFollowees(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	// todo 写 Cache

	return total, ids, nil
}

// CreateComment 创建评论
func (repo *interactiveRepository) CreateComment(ctx context.Context, comment domain.Comment, events ...*event.OutboxEvent) (domain.Comment, error) {
	m := toCommentModel(comment)
	c, err := repo.dao.CreateComment(ctx, &m, events...)
	if err != nil {
		return domain.Comment{}, toRepositoryErr(err)
	}

	return domain.Comment{
		ID:        c.ID,
		PostID:    c.PostID,
		ParentID:  c.ParentID,
		ReplyID:   c.ReplyID,
		UserID:    c.UserID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
	}, nil
}

// GetCommentByID 根据 ID 获取评论
func (repo *interactiveRepository) GetCommentByID(ctx context.Context, id int64) (domain.Comment, error) {
	comment, err := repo.dao.GetCommentByID(ctx, id)
	if err != nil {
		return domain.Comment{}, toRepositoryErr(err)
	}
	res := domain.ToCommentDomain(comment)
	return res, nil
}

// DelComment 删除评论
func (repo *interactiveRepository) DelComment(ctx context.Context, id int64, buildEvents func(cnt int) []*event.OutboxEvent) (int, error) {
	cnt, err := repo.dao.DelComment(ctx, id, buildEvents)
	if err != nil {
		return cnt, toRepositoryErr(err)
	}

	return cnt, nil
}

// GetCommentByPostID 根据帖子 ID 按页获取主评论
func (repo *interactiveRepository) GetCommentByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []domain.Comment, error) {
	total, comments, err := repo.dao.GetCommentByPostID(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	res := make([]domain.Comment, 0, len(comments))
	for _, comment := range comments {
		c := domain.ToCommentDomain(comment)
		res = append(res, c)
	}

	return total, res, nil
}

// GetCommentRepliesByParentID 根据主评论 ID 按页获取子评论
func (repo *interactiveRepository) GetCommentRepliesByParentID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []domain.Comment, error) {
	total, comments, err := repo.dao.GetCommentRepliesByParentID(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	res := make([]domain.Comment, 0, len(comments))
	for _, comment := range comments {
		c := domain.ToCommentDomain(comment)
		res = append(res, c)
	}

	return total, res, nil
}
func toCommentModel(comment domain.Comment) model.Comment {

	return model.Comment{
		ID:       comment.ID,
		PostID:   comment.PostID,
		ParentID: comment.ParentID,
		ReplyID:  comment.ReplyID,
		UserID:   comment.UserID,
		Content:  comment.Content,
	}
}
