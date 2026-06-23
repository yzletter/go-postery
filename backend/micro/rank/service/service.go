package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
	inter_grpc "github.com/yzletter/go-postery/api/proto/interactive/v1"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/errs"
	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/rank/domain"
	"github.com/yzletter/go-postery/backend/micro/rank/repository"
	"golang.org/x/sync/errgroup"
)

type rankService struct {
	rankRepo    repository.RankRepository
	userClient  manager.UserClient
	postClient  manager.PostClient
	interClient manager.InteractiveClient
	consumer    *kafka.Reader
}

// NewRankService 构造函数
func NewRankService(rankRepo repository.RankRepository, userClient manager.UserClient, postClient manager.PostClient, interClient manager.InteractiveClient, consumer *kafka.Reader) RankService {
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
			slog.Error("get profile empty", "user_id", id)
			return errs.ErrInternal
		}

		tt, err := time.Parse(time.RFC3339, profile.CreatedAt)
		if err != nil {
			slog.Error("parse time failed", "error", err)
			return errs.ErrInternal
		}
		create = tt.Unix()
		return nil
	})

	eg.Go(func() error {
		// 获取关注
		userInter, err := svc.interClient.GetUserInteractive(ctx, &inter_grpc.UserIDRequest{UserID: id})
		if err != nil {
			slog.Error("get user interactive failed", "error", err, "user_id", id)
			return err
		}
		if userInter == nil {
			slog.Error("get user interactive empty", "user_id", id)
			return errs.ErrInternal
		}
		follow = userInter.FollowCnt
		return nil
	})

	if err := eg.Wait(); err != nil {
		slog.Error("get user dimension failed")
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
			slog.Error("get post by id failed", "error", eerr)
			return errs.ErrInternal
		}
		if post == nil {
			slog.Error("get post by id empty", "post_id", id)
			return errs.ErrInternal
		}
		tt, eerr := time.Parse(time.RFC3339, post.CreatedAt)
		if eerr != nil {
			slog.Error("parse time failed", "error", eerr)
			return errs.ErrInternal
		}
		create = tt.Unix()
		return nil
	})

	eg.Go(func() error {
		// 获取喜欢、不喜欢、评论数
		postInter, eerr := svc.interClient.GetPostInteractive(ctx, &inter_grpc.PostIDRequest{PostID: id})
		if eerr != nil {
			slog.Error("get post interactive failed", "error", eerr)
			return errs.ErrInternal
		}
		if postInter == nil {
			slog.Error("get post interactive empty", "post_id", id)
			return errs.ErrInternal
		}
		like, comment = postInter.LikeCnt, postInter.CommentCnt
		return nil
	})

	if err := eg.Wait(); err != nil {
		slog.Error("get post dimension failed")
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

// RankTopKUser 计算用户榜单
func (svc *rankService) RankTopKUser(ctx context.Context) error {
	// 获取七天内用户 ID
	resp, err := svc.userClient.GetIDAfterTime(ctx, &user_grpc.GetIDAfterTimeRequest{
		TimeAfter: time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
	})
	if err != nil {
		slog.Error("get user by time failed", "error", err)
		return errs.ErrInternal
	}
	if resp == nil {
		slog.Error("get user by time empty")
		return errs.ErrInternal
	}

	// 计算分数
	ids := resp.IDs
	for _, id := range ids {
		if err := svc.RankUser(ctx, id); err != nil {
			slog.Error("rank user failed", "error", err, "id", id)
		}
	}
	return nil
}

// RankTopKPost 计算文章榜单
func (svc *rankService) RankTopKPost(ctx context.Context) error {
	// 获取七天内文章 ID
	resp, err := svc.postClient.GetPostByTime(ctx, &post_grpc.GetPostByTimeRequest{
		TimeAt: time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
	})
	if err != nil {
		slog.Error("get post by time failed", "error", err)
		return errs.ErrInternal
	}
	if resp == nil {
		slog.Error("get post by time empty")
		return errs.ErrInternal
	}

	// 计算分数
	ids := resp.IDs
	for _, id := range ids {
		if err := svc.RankPost(ctx, id); err != nil {
			slog.Error("rank post failed", "error", err, "id", id)
		}
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

func (svc *rankService) CronRankTopK() {
	ctx := context.Background()
	go svc.RankTopKPost(ctx)
	go svc.RankTopKUser(ctx)
}

func (svc *rankService) StartKafkaConsumer(ctx context.Context) {
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			slog.Info("close kafka consumer goroutine")
			return
		default:
			msg, err := svc.consumer.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				slog.Error("server internal error", "error", err)

				// 失败进行退避
				time.Sleep(backoff)
				if backoff < 10*time.Second {
					backoff *= 2
				}
				continue
			}

			// 重置退避
			backoff = time.Second

			// 反序列化消息
			var e event.UpdateEvent
			if err := sonic.Unmarshal(msg.Value, &e); err != nil {
				slog.Error("unmarshal event failed", "error", err)
				// 脏消息 Commit 掉
				_ = svc.consumer.CommitMessages(ctx, msg)
				continue
			}

			// 判断类型, 计算分数
			switch e.Biz {
			case event.UpdateUserScore:
				if err := svc.RankUser(ctx, e.BizID); err != nil {
					slog.Error("rank user failed", "error", err, "biz_id", e.BizID)
				}
			case event.UpdatePostScore:
				if err := svc.RankPost(ctx, e.BizID); err != nil {
					slog.Error("rank post failed", "error", err, "biz_id", e.BizID)
				}
			default:
				slog.Error("unknown rank update biz", "biz", e.Biz, "biz_id", e.BizID)
			}

			// 将消息全都 Commit 不管成功失败
			if err := svc.consumer.CommitMessages(ctx, msg); err != nil {
				slog.Error("commit messages failed", "error", err)
				continue
			}
		}
	}
}
