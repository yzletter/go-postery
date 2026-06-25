package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
	"github.com/yzletter/go-postery/backend/micro/interactive/repository"
)

// Like 点赞
func (svc *interactiveService) Like(ctx context.Context, userID int64, postID int64) error {
	// 查询帖子
	resp, err := svc.postClient.ExistPost(ctx, &post_grpc.ExistPostRequest{PostID: postID})
	if err != nil || resp == nil {
		slog.Error("check post before like failed", "post_id", postID, "error", err)
		return errs.ErrInternal
	}

	if resp.Exist == false {
		slog.Info("like rejected: post not found", "post_id", postID)
		return errs.ErrNotFound
	}

	now := time.Now()

	// 创建点赞记录
	like := domain.Like{
		ID:     svc.idGen.NextID(),
		UserID: userID,
		PostID: postID,
	}

	// 消息
	e := event.NewLikeEventPayload{
		ID:       svc.idGen.NextID(),
		UserID:   userID,
		PostID:   postID,
		LikeType: event.Like,
		EventAt:  now,
	}

	ee := event.UpdateScoreEventPayload{
		ID:    svc.idGen.NextID(),
		Biz:   event.UpdatePostScore,
		BizID: postID,
	}
	if err := svc.interRepo.Like(ctx, like,
		event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicInteractiveLike, event.KafkaInteractiveGroup, e),
		event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicRankUpdateScore, event.KafkaRankGroup, ee),
	); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			slog.Info("like skipped: already liked", "user_id", userID, "post_id", postID)
			return errs.ErrAlreadyExits
		}
		// 系统内部错误
		slog.Error("create like failed", "user_id", userID, "post_id", postID, "error", err)
		return errs.ErrInternal
	}

	return nil
}

// Unlike 取消点赞
func (svc *interactiveService) Unlike(ctx context.Context, postID int64, userID int64) error {
	// 查询帖子
	resp, err := svc.postClient.ExistPost(ctx, &post_grpc.ExistPostRequest{PostID: postID})
	if err != nil || resp == nil {
		slog.Error("check post before unlike failed", "post_id", postID, "error", err)
		return errs.ErrInternal
	}

	if resp.Exist == false {
		slog.Info("unlike rejected: post not found", "post_id", postID)
		return errs.ErrNotFound
	}

	now := time.Now()

	e := event.NewLikeEventPayload{
		ID:       svc.idGen.NextID(),
		UserID:   userID,
		PostID:   postID,
		LikeType: event.Unlike,
		EventAt:  now,
	}

	ee := event.UpdateScoreEventPayload{
		ID:    svc.idGen.NextID(),
		Biz:   event.UpdatePostScore,
		BizID: postID,
	}

	// 删除点赞记录
	if err := svc.interRepo.UnLike(ctx, userID, postID,
		event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicInteractiveLike, event.KafkaInteractiveGroup, e),
		event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicRankUpdateScore, event.KafkaRankGroup, ee),
	); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("unlike skipped: not liked", "user_id", userID, "post_id", postID)
			return errs.ErrAlreadyExits
		}
		// 系统内部错误
		slog.Error("delete like failed", "user_id", userID, "post_id", postID, "error", err)
		return errs.ErrInternal
	}

	return nil
}

// CheckLike 用户是否已点赞
func (svc *interactiveService) CheckLike(ctx context.Context, userID int64, postID int64) (bool, error) {
	res, err := svc.interRepo.HasLiked(ctx, userID, postID)
	if err != nil {
		slog.Error("check like failed", "user_id", userID, "post_id", postID, "error", err)
		return false, errs.ErrInternal
	}
	return res, nil
}
