package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/event/outbox/model"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
	"github.com/yzletter/go-postery/backend/micro/interactive/repository"
	"github.com/yzletter/go-postery/backend/ports"
)

type interactiveService struct {
	interRepo repository.InteractiveRepository

	postClient manager.PostClient
	idGen      ports.IDGenerator
	consumer   *kafka.Reader
}

func NewInteractiveService(interRepo repository.InteractiveRepository,
	postClient manager.PostClient, idGen ports.IDGenerator, consumer *kafka.Reader) InteractiveService {
	return &interactiveService{
		interRepo:  interRepo,
		postClient: postClient,
		idGen:      idGen,
		consumer:   consumer,
	}
}

// GetPostInteractive 获取帖子互动信息
func (svc *interactiveService) GetPostInteractive(ctx context.Context, id int64) (domain.PostInter, error) {
	postInter, err := svc.interRepo.GetPostInteractive(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("post interactive not found", "id", id)
			return domain.PostInter{}, nil
		}
		slog.Error("get post interactive failed", "id", id, "error", err)
		return domain.PostInter{}, errs.ErrInternal
	}
	return postInter, nil
}

// GetUserInteractive 获取用户互动信息
func (svc *interactiveService) GetUserInteractive(ctx context.Context, id int64) (domain.UserInter, error) {
	userInter, err := svc.interRepo.GetUserInteractive(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("user interactive not found", "id", id)
			return domain.UserInter{}, nil
		}
		slog.Error("get user interactive failed", "id", id, "error", err)
		return domain.UserInter{}, errs.ErrInternal
	}
	return userInter, nil
}

const (
	batch        = 50
	flushTimeout = 5 * time.Second
)

// StartKafkaConsumer 启动互动消息消费者
func (svc *interactiveService) StartKafkaConsumer(ctx context.Context) {
	backoff := time.Second
	readMsgs := make([]kafka.Message, 0, batch)

	flush := func(ctx context.Context) {
		if len(readMsgs) == 0 {
			return
		}
		if err := svc.batchConsumeReadMessage(ctx, readMsgs); err != nil {
			return
		}
		readMsgs = readMsgs[:0]
	}

	for {
		// 带超时进行 Fetch
		fetchCtx, cancel := context.WithTimeout(ctx, flushTimeout)
		msg, err := svc.consumer.FetchMessage(fetchCtx)
		cancel()
		if err != nil {
			// 消费者退出
			if ctx.Err() != nil {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				// 退出前 flush
				flush(shutdownCtx)
				cancel()

				slog.Info("close interactive event consumer success")
				return
			}

			// Fetch 超时, 很久没有消息
			if errors.Is(err, context.DeadlineExceeded) {
				// flush
				flush(ctx)
				continue
			}

			// 其他错误进行退避
			slog.Warn("fetch interactive event failed", "error", err)
			time.Sleep(backoff)
			if backoff < 10*time.Second {
				backoff *= 2
			}
			continue
		}

		// 重置退避
		backoff = time.Second

		switch msg.Topic {
		// 阅读消息
		case event.KafkaTopicInteractiveRead:
			// 积攒消息
			readMsgs = append(readMsgs, msg)
			if len(readMsgs) >= batch {
				// 积攒够了进行 flush
				flush(ctx)
			}
		// 点赞消息
		case event.KafkaTopicInteractiveLike:
			// 反序列化消息
			var e event.NewLikeEventPayload
			if err := sonic.Unmarshal(msg.Value, &e); err != nil {
				slog.Warn("invalid like event, skip", "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset, "error", err)
				// 脏消息 Commit 掉
				_ = svc.consumer.CommitMessages(ctx, msg)
				continue
			}

			// 判断 delta
			delta := 1
			if e.LikeType != event.Like {
				delta = -1
			}

			processedEvent := &model.ProcessedEvent{
				ID:       svc.idGen.NextID(),
				Consumer: event.KafkaInteractiveGroup,
				EventID:  e.ID,
				Topic:    msg.Topic,
			}

			// 消费消息
			if err := svc.interRepo.ChangeInteractiveCntWithOutbox(ctx, domain.BizLike, e.PostID, e.EventAt, int64(delta), processedEvent); err != nil {
				slog.Error("change like count failed", "post_id", e.PostID, "delta", delta, "error", err)
				continue
			}

			// Commit 消息
			if err := svc.consumer.CommitMessages(ctx, msg); err != nil {
				slog.Error("commit like event failed", "post_id", e.PostID, "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset, "error", err)
				continue
			}

		// 关注消息
		case event.KafkaTopicInteractiveFollow:
			// 反序列化消息
			var e event.NewFollowEventPayload
			if err := sonic.Unmarshal(msg.Value, &e); err != nil {
				slog.Warn("invalid follow event, skip", "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset, "error", err)
				// 脏消息 Commit 掉
				_ = svc.consumer.CommitMessages(ctx, msg)
				continue
			}

			// 判断 delta
			delta := 1
			if e.FollowType != event.Follow {
				delta = -1
			}

			processedEvent := &model.ProcessedEvent{
				ID:       svc.idGen.NextID(),
				Consumer: event.KafkaInteractiveGroup,
				EventID:  e.ID,
				Topic:    msg.Topic,
			}

			// 消费消息
			if err := svc.interRepo.ChangeInteractiveCntWithOutbox(ctx, domain.BizFollow, e.Followee, e.EventAt, int64(delta), processedEvent); err != nil {
				slog.Error("change follow count failed", "followee", e.Followee, "delta", delta, "error", err)
				continue
			}

			// Commit 消息
			if err := svc.consumer.CommitMessages(ctx, msg); err != nil {
				slog.Error("commit follow event failed", "followee", e.Followee, "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset, "error", err)
				continue
			}
		// 评论消息
		case event.KafkaTopicInteractiveComment:
			// 反序列化消息
			var e event.NewCommentEventPayload
			if err := sonic.Unmarshal(msg.Value, &e); err != nil {
				slog.Warn("invalid comment event, skip", "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset, "error", err)
				// 脏消息 Commit 掉
				_ = svc.consumer.CommitMessages(ctx, msg)
				continue
			}

			// 判断 delta
			delta := e.Cnt
			if e.CommentType != event.Create {
				// 删评论
				delta = -e.Cnt
			}

			processedEvent := model.ProcessedEvent{
				ID:       svc.idGen.NextID(),
				Consumer: event.KafkaInteractiveGroup,
				EventID:  e.ID,
				Topic:    msg.Topic,
			}

			// 消费消息
			if err := svc.interRepo.ChangeInteractiveCntWithOutbox(ctx, domain.BizComment, e.PostID, e.EventAt, int64(delta), &processedEvent); err != nil {
				slog.Error("change comment count failed", "post_id", e.PostID, "delta", delta, "error", err)
				continue
			}

			// Commit 消息
			if err := svc.consumer.CommitMessages(ctx, msg); err != nil {
				slog.Error("commit comment event failed", "post_id", e.PostID, "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset, "error", err)
				continue
			}
		}

	}
}

// 批量消费阅读消息
func (svc *interactiveService) batchConsumeReadMessage(ctx context.Context, readMsgs []kafka.Message) error {
	// 聚集 payloads
	payloads := make([]*event.NewReadEventPayload, 0)
	for _, readMsg := range readMsgs {
		var e event.NewReadEventPayload
		if err := sonic.Unmarshal(readMsg.Value, &e); err != nil {
			slog.Warn("invalid read event, skip", "topic", readMsg.Topic, "partition", readMsg.Partition, "offset", readMsg.Offset, "error", err)
			continue
		}
		payloads = append(payloads, &e)
	}

	// 消费
	if err := svc.interRepo.IncrReadCnt(ctx, event.KafkaInteractiveGroup, event.KafkaTopicInteractiveRead, payloads...); err != nil {
		return err
	}

	// 批量 Commit
	if err := svc.consumer.CommitMessages(ctx, readMsgs...); err != nil {
		slog.Error("commit read events failed", "count", len(readMsgs), "error", err)
		return err
	}

	return nil
}
