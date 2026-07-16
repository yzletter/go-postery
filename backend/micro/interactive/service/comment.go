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

// Comment 评论
func (svc *interactiveService) Comment(ctx context.Context, postID int64, parentID int64, replyID int64, userID int64, content string) (domain.Comment, error) {
	// 查询帖子是否存在
	resp, err := svc.postClient.ExistPost(ctx, &post_grpc.ExistPostRequest{PostID: postID})
	if err != nil || resp == nil {
		slog.Error("check post before comment failed", "post_id", postID, "error", err)
		return domain.Comment{}, errs.ErrInternal
	}
	if resp.Exist == false {
		slog.Info("comment rejected: post not found", "post_id", postID)
		return domain.Comment{}, errs.ErrNotFound
	}

	if parentID == 0 && replyID > 0 {
		return domain.Comment{}, errs.ErrInvalidArgument
	}

	// 校验主评论、所回复的评论
	if parentID > 0 {
		parent, err := svc.interRepo.GetCommentByID(ctx, parentID)
		if err != nil {
			// 主评论不存在
			if errors.Is(err, repository.ErrRecordNotFound) {
				slog.Error("parent comment not found", "parent_id", parentID)
				return domain.Comment{}, errs.ErrNotFound
			}
			slog.Error("get reply comment failed", "reply_id", replyID, "error", err)
			return domain.Comment{}, errs.ErrInternal
		} else if parent.PostID != postID {
			// 不属于一个帖子
			slog.Error("parent comment does not match post", "post_id", postID)
			return domain.Comment{}, errs.ErrInternal
		}
	}
	if replyID > 0 {
		reply, err := svc.interRepo.GetCommentByID(ctx, replyID)
		if err != nil {
			// 评论不存在
			if errors.Is(err, repository.ErrRecordNotFound) {
				slog.Error("reply comment not found", "parent_id", parentID)
				return domain.Comment{}, errs.ErrNotFound
			}
			slog.Error("get reply comment failed", "reply_id", replyID, "error", err)
			return domain.Comment{}, errs.ErrInternal
		} else if reply.PostID != postID {
			// 不属于一个帖子
			slog.Error("reply comment does not match post", "post_id", postID)
			return domain.Comment{}, errs.ErrInternal
		}
	}

	now := time.Now()

	// 新建评论
	comment := domain.Comment{
		ID:       svc.idGen.NextID(),
		PostID:   postID,
		ParentID: parentID,
		ReplyID:  replyID,
		UserID:   userID,
		Content:  content,
	}

	e := event.NewCommentEventPayload{
		ID:          svc.idGen.NextID(),
		UserID:      userID,
		PostID:      postID,
		Cnt:         1,
		CommentType: event.Create,
		EventAt:     now,
	}

	ee := event.UpdateScoreEventPayload{
		ID:    svc.idGen.NextID(),
		Biz:   event.UpdatePostScore,
		BizID: comment.PostID,
	}
	if comment, err = svc.interRepo.CreateComment(ctx, comment,
		event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicInteractiveComment, event.KafkaInteractiveGroup, e),
		event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicRankUpdateScore, event.KafkaRankGroup, ee),
	); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			slog.Warn("create comment id conflict", "comment_id", comment.ID, "post_id", postID, "error", err)
		} else {
			slog.Error("create comment failed", "post_id", postID, "user_id", userID, "error", err)
		}
		return domain.Comment{}, errs.ErrInternal
	}

	return comment, nil
}

func (svc *interactiveService) DelComment(ctx context.Context, commentID int64, userID int64) error {
	// 查找评论
	comment, err := svc.interRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("delete comment skipped: comment not found", "comment_id", commentID)
			return errs.ErrNotFound
		}
		slog.Error("get comment before delete failed", "comment_id", commentID, "error", err)
		return errs.ErrInternal
	}

	// 判断是否有删除权限
	ok, err := svc.CheckCommentDelAuth(ctx, commentID, userID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			slog.Error("delete comment skipped: comment not found", "comment_id", commentID)
			return errs.ErrNotFound
		}
		slog.Error("check comment delete auth failed", "comment_id", commentID, "error", err)
		return errs.ErrInternal
	} else if !ok {
		// 没有权限删除
		slog.Info("delete comment rejected: unauthenticated", "user_id", userID, "comment_id", commentID)
		return errs.ErrUnauthenticated
	}

	now := time.Now()

	ee := event.UpdateScoreEventPayload{
		ID:    svc.idGen.NextID(),
		Biz:   event.UpdatePostScore,
		BizID: comment.PostID,
	}

	buildEventsFunc := func(cnt int) []*event.OutboxEvent {
		e := event.NewCommentEventPayload{
			ID:          svc.idGen.NextID(),
			UserID:      userID,
			PostID:      comment.PostID,
			Cnt:         cnt,
			CommentType: event.Del,
			EventAt:     now,
		}

		return []*event.OutboxEvent{
			event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicInteractiveComment, event.KafkaInteractiveGroup, e),
			event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaTopicRankUpdateScore, event.KafkaRankGroup, ee),
		}
	}

	// 删除评论
	if _, err := svc.interRepo.DelComment(ctx, commentID, buildEventsFunc); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("delete comment skipped: comment not found", "comment_id", commentID)
			return errs.ErrNotFound
		}
		slog.Error("delete comment failed", "comment_id", commentID, "error", err)
		return errs.ErrInternal
	}
	return nil
}

// ListCommentByPage 根据 PostID 按页获取文章主评论
func (svc *interactiveService) ListCommentByPage(ctx context.Context, postID int64, pageNo int, pageSize int) (int64, []domain.Comment, error) {
	total, comments, err := svc.interRepo.GetCommentByPostID(ctx, postID, pageNo, pageSize)
	if err != nil {
		if errors.Is(err, repository.ErrParamsInvalid) {
			slog.Info("list comments rejected: invalid page params", "post_id", postID, "page_no", pageNo, "page_size", pageSize)
			return 0, nil, errs.ErrInvalidArgument
		}
		slog.Error("list comments by post failed", "post_id", postID, "page_no", pageNo, "page_size", pageSize, "error", err)
		return 0, nil, errs.ErrInternal
	}

	return total, comments, nil
}

// ListRepliesByPage 根据 CommentID 按页获取评论的回复
func (svc *interactiveService) ListRepliesByPage(ctx context.Context, commentID int64, pageNo int, pageSize int) (int64, []domain.Comment, error) {
	total, comments, err := svc.interRepo.GetCommentRepliesByParentID(ctx, commentID, pageNo, pageSize)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("list replies skipped: parent comment not found", "comment_id", commentID)
			return 0, nil, errs.ErrNotFound
		}
		if errors.Is(err, repository.ErrParamsInvalid) {
			slog.Info("list replies rejected: invalid page params", "comment_id", commentID, "page_no", pageNo, "page_size", pageSize)
			return 0, nil, errs.ErrInvalidArgument
		}
		slog.Error("list comment replies failed", "comment_id", commentID, "page_no", pageNo, "page_size", pageSize, "error", err)
		return 0, nil, errs.ErrInternal
	}
	return total, comments, nil
}

// CheckCommentDelAuth 用户是否有删除权限
func (svc *interactiveService) CheckCommentDelAuth(ctx context.Context, commentID int64, userID int64) (bool, error) {
	// 查询评论是否存在
	comment, err := svc.interRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("check comment auth skipped: comment not found", "comment_id", commentID)
			return false, errs.ErrNotFound
		}
		slog.Error("get comment for auth check failed", "comment_id", commentID, "error", err)
		return false, errs.ErrInternal
	}

	// 查询用户是否有帖子权限
	resp, err := svc.postClient.CheckPostAuth(ctx, &post_grpc.CheckPostAuthRequest{PostID: comment.PostID, UserID: userID})
	if err != nil || resp == nil {
		slog.Error("check post auth for comment failed", "post_id", comment.PostID, "user_id", userID, "error", err)
		return false, errs.ErrInternal
	}

	// 用户有帖子权限，或评论属于该用户
	return comment.UserID == userID || resp.Exist, nil
}
