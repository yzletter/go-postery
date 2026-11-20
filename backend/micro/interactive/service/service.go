package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
	"github.com/yzletter/go-postery/backend/micro/interactive/repository"
	"github.com/yzletter/go-postery/backend/micro/outbox/model"
	outboxService "github.com/yzletter/go-postery/backend/micro/outbox/service"
	"github.com/yzletter/go-postery/backend/ports"
)

type interactiveService struct {
	interRepo repository.InteractiveRepository

	postClient manager.PostClient
	idGen      ports.IDGenerator
	consumer   *outboxService.Consumer
}

func NewInteractiveService(interRepo repository.InteractiveRepository,
	postClient manager.PostClient, idGen ports.IDGenerator, consumer *outboxService.Consumer) InteractiveService {
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

const readBatchSize = 50

// StartKafkaConsumer 启动互动消息消费者。
func (svc *interactiveService) StartKafkaConsumer(ctx context.Context) {
	svc.consumer.Run(ctx, svc.consumeMessages, map[string]int{model.KafkaTopicInteractiveRead: readBatchSize})
}

// consumeMessages 处理互动事件。阅读事件按批处理，其他事件逐条更新计数
func (svc *interactiveService) consumeMessages(ctx context.Context, messages ...kafka.Message) error {
	msg := messages[0]
	if msg.Topic == model.KafkaTopicInteractiveRead {
		return svc.consumeReadMessages(ctx, messages)
	}

	// 使用原消息中的事件 ID 去重，重试和回放时都沿用该 ID
	processedEvent := &model.ProcessedEvent{
		ID:       svc.idGen.NextID(),
		Consumer: model.KafkaInteractiveGroup,
		Topic:    msg.Topic,
	}

	switch msg.Topic {
	case model.KafkaTopicInteractiveLike:
		var e model.NewLikeEventPayload
		if err := outboxService.DecodeMessage(msg, &e); err != nil {
			return err
		}
		if e.ID <= 0 || e.PostID <= 0 || (e.LikeType != model.Like && e.LikeType != model.Unlike) {
			return fmt.Errorf("%w: invalid like event", outboxService.ErrInvalidMessage)
		}
		delta := int64(1)
		if e.LikeType == model.Unlike {
			delta = -1
		}
		processedEvent.EventID = e.ID
		return svc.interRepo.ChangeInteractiveCntWithOutbox(ctx, domain.BizLike, e.PostID, e.EventAt, delta, processedEvent)
	case model.KafkaTopicInteractiveFollow:
		var e model.NewFollowEventPayload
		if err := outboxService.DecodeMessage(msg, &e); err != nil {
			return err
		}
		if e.ID <= 0 || e.Followee <= 0 || (e.FollowType != model.Follow && e.FollowType != model.Unfollow) {
			return fmt.Errorf("%w: invalid follow event", outboxService.ErrInvalidMessage)
		}
		delta := int64(1)
		if e.FollowType == model.Unfollow {
			delta = -1
		}
		processedEvent.EventID = e.ID
		return svc.interRepo.ChangeInteractiveCntWithOutbox(ctx, domain.BizFollow, e.Followee, e.EventAt, delta, processedEvent)
	case model.KafkaTopicInteractiveComment:
		var e model.NewCommentEventPayload
		if err := outboxService.DecodeMessage(msg, &e); err != nil {
			return err
		}
		if e.ID <= 0 || e.PostID <= 0 || e.Cnt <= 0 || (e.CommentType != model.Create && e.CommentType != model.Del) {
			return fmt.Errorf("%w: invalid comment event", outboxService.ErrInvalidMessage)
		}
		delta := int64(e.Cnt)
		if e.CommentType == model.Del {
			delta = -delta
		}
		processedEvent.EventID = e.ID
		return svc.interRepo.ChangeInteractiveCntWithOutbox(ctx, domain.BizComment, e.PostID, e.EventAt, delta, processedEvent)
	default:
		return fmt.Errorf("%w: unexpected interactive topic %q", outboxService.ErrInvalidMessage, msg.Topic)
	}
}

// consumeReadMessages 批量更新阅读数，同一事务中对事件去重。
func (svc *interactiveService) consumeReadMessages(ctx context.Context, messages []kafka.Message) error {
	payloads := make([]*model.NewReadEventPayload, 0, len(messages))
	for _, msg := range messages {
		var e model.NewReadEventPayload
		if err := outboxService.DecodeMessage(msg, &e); err != nil {
			return err
		}
		if e.ID <= 0 || e.PostID <= 0 {
			return fmt.Errorf("%w: invalid read event", outboxService.ErrInvalidMessage)
		}
		payloads = append(payloads, &e)
	}
	return svc.interRepo.IncrReadCnt(ctx, model.KafkaInteractiveGroup, model.KafkaTopicInteractiveRead, payloads...)
}
