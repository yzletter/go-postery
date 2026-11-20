package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	inter_grpc "github.com/yzletter/go-postery/backend/api/proto/interactive/v1"
	"github.com/yzletter/go-postery/backend/api/proto/post/v1"
	"github.com/yzletter/go-postery/backend/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/outbox/model"
	outboxService "github.com/yzletter/go-postery/backend/micro/outbox/service"
	"github.com/yzletter/go-postery/backend/micro/rank/domain"
	"github.com/yzletter/go-postery/backend/micro/rank/repository"
	"golang.org/x/sync/errgroup"
)

type rankService struct {
	rankRepo    repository.RankRepository
	userClient  manager.UserClient
	postClient  manager.PostClient
	interClient manager.InteractiveClient
	consumer    *outboxService.Consumer
}

// NewRankService 构造函数
func NewRankService(rankRepo repository.RankRepository, userClient manager.UserClient, postClient manager.PostClient, interClient manager.InteractiveClient, consumer *outboxService.Consumer) RankService {
	return &rankService{
		rankRepo:    rankRepo,
		userClient:  userClient,
		postClient:  postClient,
		interClient: interClient,
		consumer:    consumer,
	}
}

// RankUser 计算用户分数并放入排行榜
func (svc *rankService) RankUser(ctx context.Context, id int64) error {
	// 分数 = 创建时间 + 关注 * 系数
	var create, follow int64
	var eg errgroup.Group
	eg.Go(func() error {
		// 获取创建时间
		profile, eerr := svc.userClient.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: id})
		if eerr != nil {
			return eerr
		}
		if profile == nil {
			return errors.New("profile is empty")
		}

		tt, err := time.Parse(time.RFC3339, profile.CreatedAt)
		if err != nil {
			return err
		}
		create = tt.Unix()
		return nil
	})

	eg.Go(func() error {
		// 获取关注
		userInter, err := svc.interClient.GetUserInteractive(ctx, &inter_grpc.UserIDRequest{UserID: id})
		if err != nil {
			return err
		}
		if userInter == nil {
			return errors.New("user interactive is empty")
		}
		follow = userInter.FollowCnt
		return nil
	})

	if err := eg.Wait(); err != nil {
		slog.Error("rank user dimensions failed", "user_id", id, "error", err)
		return errs.ErrInternal
	}

	// 计算分数
	score := create + follow*domain.FollowCoefficient

	// 放入缓存
	if err := svc.rankRepo.UpdateUserScore(ctx, id, score); err != nil {
		slog.Error("rank user failed", "error", err, "id", id)
		return errs.ErrInternal
	}
	return nil
}

// RankPost 计算文章分数并放入排行榜
func (svc *rankService) RankPost(ctx context.Context, id int64) error {
	// 分数 = 创建时间 + 喜欢 * 系数 + 评论 * 系数
	var create, like, comment int64
	var eg errgroup.Group
	eg.Go(func() error {
		post, eerr := svc.postClient.GetBriefByID(ctx, &post_grpc.GetBriefByIDRequest{PostID: id})
		if eerr != nil {
			return eerr
		}
		if post == nil {
			return errors.New("post is empty")
		}
		tt, eerr := time.Parse(time.RFC3339, post.CreatedAt)
		if eerr != nil {
			return eerr
		}
		create = tt.Unix()
		return nil
	})

	eg.Go(func() error {
		// 获取喜欢、不喜欢、评论数
		postInter, eerr := svc.interClient.GetPostInteractive(ctx, &inter_grpc.PostIDRequest{PostID: id})
		if eerr != nil {
			return eerr
		}
		if postInter == nil {
			return errors.New("post interactive is empty")
		}
		like, comment = postInter.LikeCnt, postInter.CommentCnt
		return nil
	})

	if err := eg.Wait(); err != nil {
		slog.Error("rank post dimensions failed", "post_id", id, "error", err)
		return errs.ErrInternal
	}

	// 计算分数
	score := create + like*domain.LikeCoefficient + comment*domain.CommentCoefficient

	// 放入缓存
	if err := svc.rankRepo.UpdatePostScore(ctx, id, score); err != nil {
		slog.Error("rank post failed", "error", err, "id", id)
		return errs.ErrInternal
	}
	return nil
}

// RankTopKUser 完整重新计算用户榜单
func (svc *rankService) RankTopKUser(ctx context.Context) error {
	// 获取七天内用户 ID
	resp, err := svc.userClient.GetIDAfterTime(ctx, &user_grpc.GetIDAfterTimeRequest{
		TimeAfter: time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
	})
	if err != nil {
		slog.Error("get recent user ids failed", "error", err)
		return errs.ErrInternal
	}
	if resp == nil {
		slog.Warn("get recent user ids returned empty response")
		return errs.ErrInternal
	}

	// 计算分数
	for _, id := range resp.IDs {
		_ = svc.RankUser(ctx, id)
	}
	return nil
}

// RankTopKPost 完整重新计算文章榜单
func (svc *rankService) RankTopKPost(ctx context.Context) error {
	// 获取七天内文章 ID
	resp, err := svc.postClient.GetPostByTime(ctx, &post_grpc.GetPostByTimeRequest{
		TimeAt: time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
	})
	if err != nil {
		slog.Error("get recent post ids failed", "error", err)
		return errs.ErrInternal
	}
	if resp == nil {
		slog.Warn("get recent post ids returned empty response")
		return errs.ErrInternal
	}

	// 计算分数
	for _, id := range resp.IDs {
		_ = svc.RankPost(ctx, id)
	}

	return nil
}

// TopKUser 返回用户榜单
func (svc *rankService) TopKUser(ctx context.Context) ([]domain.User, error) {
	users, err := svc.rankRepo.GetTopKUser(ctx)
	if err != nil {
		slog.Error("get top user failed", "error", err)
		return nil, errs.ErrInternal
	}
	return users, nil
}

// TopKPost 返回文章榜单
func (svc *rankService) TopKPost(ctx context.Context) ([]domain.Post, error) {
	posts, err := svc.rankRepo.GetTopKPost(ctx)
	if err != nil {
		slog.Error("get top post failed", "error", err)
		return nil, errs.ErrInternal
	}
	return posts, nil
}

// DeleteUserScore 删除用户分数
func (svc *rankService) DeleteUserScore(ctx context.Context, id int64) error {
	if err := svc.rankRepo.DeleteUserScore(ctx, id); err != nil {
		slog.Error("delete user score failed", "id", id, "error", err)
		return errs.ErrInternal
	}
	return nil
}

// DeletePostScore 删除帖子分数
func (svc *rankService) DeletePostScore(ctx context.Context, id int64) error {
	if err := svc.rankRepo.DeletePostScore(ctx, id); err != nil {
		slog.Error("delete post score failed", "id", id, "error", err)
		return errs.ErrInternal
	}
	return nil
}

// CronRankTopK 定时计算榜单
func (svc *rankService) CronRankTopK() {
	ctx := context.Background()
	go svc.RankTopKPost(ctx)
	go svc.RankTopKUser(ctx)
}

func (svc *rankService) StartKafkaConsumer(ctx context.Context) {
	svc.consumer.Run(ctx, svc.consumeScoreEvent, nil)
}

// consumeScoreEvent 更新或删除榜单分数，处理失败交给消费者重试
func (svc *rankService) consumeScoreEvent(ctx context.Context, messages ...kafka.Message) error {
	var e model.UpdateScoreEventPayload
	if err := outboxService.DecodeMessage(messages[0], &e); err != nil {
		return err
	}
	if e.BizID <= 0 {
		return fmt.Errorf("%w: invalid rank business id", outboxService.ErrInvalidMessage)
	}

	switch e.Biz {
	case model.UpdateUserScore:
		return svc.RankUser(ctx, e.BizID)
	case model.UpdatePostScore:
		return svc.RankPost(ctx, e.BizID)
	case model.DeleteUserScore:
		return svc.DeleteUserScore(ctx, e.BizID)
	case model.DeletePostScore:
		return svc.DeletePostScore(ctx, e.BizID)
	default:
		return fmt.Errorf("%w: unknown rank event type %d", outboxService.ErrInvalidMessage, e.Biz)
	}
}
