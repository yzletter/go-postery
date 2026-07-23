package service

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
	"github.com/yzletter/go-postery/backend/micro/interactive/repository"
)

// Follow 关注用户
// todo 查询用户存在性
func (svc *interactiveService) Follow(ctx context.Context, follower int64, followee int64) error {
	res, err := svc.interRepo.GetFollow(ctx, follower, followee)
	if err != nil {
		slog.Error("get follow relation failed", "follower", follower, "followee", followee, "error", err)
		return errs.ErrInternal
	}

	if res == domain.FollowIFollow || res == domain.FollowMutual { // 已经关注过了
		slog.Info("follow skipped: already followed", "follower", follower, "followee", followee)
		return errs.ErrAlreadyExits
	}

	now := time.Now()

	// 创建关注记录
	follow := domain.Follow{
		ID:         svc.idGen.NextID(),
		FollowerID: follower,
		FolloweeID: followee,
	}

	e := event.NewFollowEventPayload{
		ID:         svc.idGen.NextID(),
		Follower:   follower,
		Followee:   followee,
		FollowType: event.Follow,
		EventAt:    now,
	}

	ee := event.UpdateScoreEventPayload{
		ID:    svc.idGen.NextID(),
		Biz:   event.UpdateUserScore,
		BizID: followee,
	}
	if err = svc.interRepo.CreateFollow(ctx, follow,
		event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicInteractiveFollow, strconv.FormatInt(e.Followee, 10), e),
		event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicRankUpdateScore, strconv.FormatInt(ee.BizID, 10), ee),
	); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) { // 检查过仍冲突
			slog.Info("follow skipped: already followed", "follower", follower, "followee", followee)
			return errs.ErrAlreadyExits
		}
		slog.Error("create follow failed", "follower", follower, "followee", followee, "error", err)
		return errs.ErrInternal
	}

	return nil
}

// Unfollow 取消关注
func (svc *interactiveService) Unfollow(ctx context.Context, follower int64, followee int64) error {
	// 获取关注关系
	res, err := svc.interRepo.GetFollow(ctx, follower, followee)
	if err != nil {
		slog.Error("get follow relation failed", "follower", follower, "followee", followee, "error", err)
		return errs.ErrInternal
	}

	// 判断关注关系
	if res == domain.FollowNone || res == domain.FollowFollowMe {
		// 只有对方关注了我，或者互不关注
		slog.Info("unfollow skipped: not followed", "follower", follower, "followee", followee)
		return errs.ErrAlreadyExits
	}

	now := time.Now()

	e := event.NewFollowEventPayload{
		ID:         svc.idGen.NextID(),
		Follower:   follower,
		Followee:   followee,
		FollowType: event.Unfollow,
		EventAt:    now,
	}

	ee := event.UpdateScoreEventPayload{
		ID:    svc.idGen.NextID(),
		Biz:   event.UpdateUserScore,
		BizID: followee,
	}

	// 删除关注关系
	if err := svc.interRepo.DelFollow(ctx, follower, followee,
		event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicInteractiveFollow, strconv.FormatInt(e.Followee, 10), e),
		event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicRankUpdateScore, strconv.FormatInt(ee.BizID, 10), ee),
	); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("unfollow skipped: not followed", "follower", follower, "followee", followee)
			return errs.ErrAlreadyExits
		}
		slog.Error("delete follow failed", "follower", follower, "followee", followee, "error", err)
		return errs.ErrInternal
	}

	return nil
}

// IfFollow 判断关注关系
func (svc *interactiveService) IfFollow(ctx context.Context, follower int64, followee int64) (int, error) {
	// 获取关注关系
	res, err := svc.interRepo.GetFollow(ctx, follower, followee)
	if err != nil {
		slog.Error("get follow relation failed", "follower", follower, "followee", followee, "error", err)
		return -1, errs.ErrInternal
	}
	return int(res), nil
}

// GetFollowers 获取关注的人
func (svc *interactiveService) GetFollowers(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []int64, error) {
	total, IDs, err := svc.interRepo.GetFollowers(ctx, userID, pageNo, pageSize)
	if err != nil {
		if errors.Is(err, repository.ErrParamsInvalid) {
			slog.Info("get followers rejected: invalid page params", "user_id", userID, "page_no", pageNo, "page_size", pageSize)
			return 0, []int64{}, errs.ErrInvalidArgument
		}
		slog.Error("get followers failed", "user_id", userID, "page_no", pageNo, "page_size", pageSize, "error", err)
		return 0, []int64{}, errs.ErrInternal
	}
	return total, IDs, nil
}

// GetFollowees 获取粉丝
func (svc *interactiveService) GetFollowees(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []int64, error) {
	total, IDs, err := svc.interRepo.GetFollowees(ctx, userID, pageNo, pageSize)
	if err != nil {
		if errors.Is(err, repository.ErrParamsInvalid) {
			slog.Info("get followees rejected: invalid page params", "user_id", userID, "page_no", pageNo, "page_size", pageSize)
			return 0, []int64{}, errs.ErrInvalidArgument
		}
		slog.Error("get followees failed", "user_id", userID, "page_no", pageNo, "page_size", pageSize, "error", err)
		return 0, []int64{}, errs.ErrInternal
	}
	return total, IDs, nil
}
