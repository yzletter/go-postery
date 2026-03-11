package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/microservice-backend/session/errs"
	model2 "github.com/yzletter/go-postery/microservice-backend/session/model"
	repository2 "github.com/yzletter/go-postery/microservice-backend/session/repository"
	"github.com/yzletter/go-postery/microservice-backend/session/service/ports"
)

type sessionService struct {
	sessionRepo   repository2.SessionRepository
	messageRepo   repository2.MessageRepository
	mqConn        *amqp.Connection
	kafkaConsumer *kafka.Reader
	idGen         ports.IDGenerator
}

func NewSessionService(sessionRepo repository2.SessionRepository, messageRepo repository2.MessageRepository, mq *amqp.Connection, consumer *kafka.Reader, idGen ports.IDGenerator) SessionService {
	return &sessionService{
		sessionRepo:   sessionRepo,
		messageRepo:   messageRepo,
		mqConn:        mq,
		kafkaConsumer: consumer,
		idGen:         idGen,
	}
}

func (svc *sessionService) StartSessionRegisterConsumer(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			slog.Info("关闭 Session Register Consumer 成功 ...")
			return
		default:
			message, err := svc.kafkaConsumer.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { // 正常退出
					return
				}

				slog.Error("Fetch Message From Kafka Failed", "Kafka", "SessionKafka", "error", err)

				// 简单退避，避免狂刷日志
				time.Sleep(backoff)
				if backoff < 10*time.Second {
					backoff *= 2
				}
				continue
			}

			backoff = time.Second
			var payload model2.RegisterSessionEventPayload
			if err := sonic.Unmarshal(message.Value, &payload); err != nil {
				// 脏消息
				slog.Error("invalid message value, skip", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "value", string(message.Value), "errs", err)
				_ = svc.kafkaConsumer.CommitMessages(ctx, message) // 把 脏消息 Commit 掉，避免卡住
				continue
			}

			slog.Info("Read Kafka Message", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "key", string(message.Key), "value", string(message.Value))

			// 进行注册, 幂等
			if err := svc.register(ctx, payload.UserID); err != nil {
				slog.Error("Register Session Failed", "error", err)
				time.Sleep(time.Second) // 最小退避，避免打爆
				continue                // 不 commit -> 重试
			}

			// 把消息 Commit 掉
			if err := svc.kafkaConsumer.CommitMessages(ctx, message); err != nil {
				slog.Error("Commit Kafka Message Failed", "uid", payload.UserID, "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "errs", err)

				// Commit 失败通常会导致重复消费，但不会丢消息，可接受
				continue
			}
		}
	}
}

func (svc *sessionService) ListByUID(ctx context.Context, userID int64) ([]*model2.Session, error) {
	sessions, err := svc.sessionRepo.ListByUid(ctx, userID)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return nil, errs.ErrInternal
	}

	return sessions, nil
}

func (svc *sessionService) GetSession(ctx context.Context, userID int64, targetID int64) (*model2.Session, error) {
	session, err := svc.sessionRepo.GetByUidAndTargetID(ctx, userID, targetID)
	if err != nil {
		// 系统层面错误
		if !errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Server Internal Error", "error", err)
			return nil, errs.ErrInternal
		}

		// 查找对方 session
		session, err := svc.sessionRepo.GetByUidAndTargetID(ctx, targetID, userID)
		if err != nil {
			// 系统层面错误
			if !errors.Is(err, repository2.ErrRecordNotFound) {
				slog.Error("Server Internal Error", "error", err)
				return nil, errs.ErrInternal
			}

			// 双边都没找到，新建会话
			ssid := svc.idGen.NextID()
			newSession1 := &model2.Session{
				ID:         svc.idGen.NextID(),
				SessionID:  ssid,
				UserID:     userID,
				TargetID:   targetID,
				TargetType: 1,
			}

			newSession2 := &model2.Session{
				ID:         svc.idGen.NextID(),
				SessionID:  ssid,
				UserID:     targetID,
				TargetID:   userID,
				TargetType: 1,
			}

			err = svc.sessionRepo.Create(ctx, newSession1)
			if err != nil {
				slog.Error("Server Internal Error", "error", err)
				return nil, errs.ErrInternal
			}

			err = svc.sessionRepo.Create(ctx, newSession2)
			if err != nil {
				slog.Error("Server Internal Error", "error", err)
				return nil, errs.ErrInternal
			}
			return newSession1, nil
		} else {
			// 对方的会话有，说明只有我单边删除，同一个 sessionID 单边新建
			ssid := session.SessionID
			newSession1 := &model2.Session{
				ID:         svc.idGen.NextID(),
				SessionID:  ssid,
				UserID:     userID,
				TargetID:   targetID,
				TargetType: 1,
			}

			err = svc.sessionRepo.Create(ctx, newSession1)
			if err != nil {
				slog.Error("Server Internal Error", "error", err)
				return nil, errs.ErrInternal
			}
			return newSession1, nil
		}
	}

	return session, nil
}

// Register 注册用户的 Exchange 和 Queue
func (svc *sessionService) register(ctx context.Context, uid int64) error {

	// 定义 Exchange 和 Queue 名字
	exchangeName := fmt.Sprintf("%d_exchange", uid)
	queueNameComputer := fmt.Sprintf("%d_computer", uid)
	queueNameMobile := fmt.Sprintf("%d_mobile", uid)
	queueNames := []string{queueNameComputer, queueNameMobile}

	ch, err := svc.mqConn.Channel()
	if err != nil {
		return errs.ErrInternal
	}
	defer ch.Close()

	// 声明 Exchange
	err = ch.ExchangeDeclare(
		exchangeName,
		"fanout", // fanout 模式
		true,     // 持久化
		false,
		false,
		false,
		nil)

	if err != nil {
		slog.Error("Exchange Declare Failed", "uid", uid)
		return errs.ErrInternal
	}

	args := amqp.Table{
		"x-message-ttl":          int32(14 * 24 * 3600 * 1000), // 消息过期 TTL
		"x-dead-letter-exchange": "dlx",                        // 过期消息丢入死信队列
	}
	for _, queueName := range queueNames {
		// 申明队列
		if _, err := ch.QueueDeclare(queueName, true, false, false, false, args); err != nil {
			slog.Error("Queue Declare Failed", "uid", uid)
			return errs.ErrInternal
		}

		// 将队列绑定到交换机
		err = ch.QueueBind(
			queueName,    // 队列名
			"",           // fanout 模式忽略 routing key
			exchangeName, // 交换机名
			false,
			nil,
		)

		if err != nil {
			slog.Error("Queue BindTag Failed", "queue_name", queueName)
			return errs.ErrInternal
		}
	}

	return nil
}

func (svc *sessionService) GetHistoryMessagesByPage(ctx context.Context, userID int64, targetID int64, pageNo int, pageSize int) (int64, []*model2.Message, error) {
	total, messages, err := svc.messageRepo.GetByPage(ctx, userID, targetID, pageNo, pageSize)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return 0, nil, errs.ErrInternal
	}

	return int64(total), messages, nil
}

func (svc *sessionService) Delete(ctx context.Context, userID int64, sessionID int64) error {
	// 查当前用户这边的会话
	session, err := svc.sessionRepo.GetByID(ctx, userID, sessionID)
	if err != nil {
		// 幂等
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Sesison Not Found", "error", err)
			return nil
		}
		// 系统层面错误
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	if session.UserID != userID {
		slog.Error("Unauthenticated")
		return errs.ErrUnauthenticated
	}

	// 删除当前用户这边的会话, 要传 uid
	if err := svc.sessionRepo.Delete(ctx, userID, sessionID); err != nil {
		// 幂等
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Sesison Not Found", "error", err)
			return nil
		}
		// 系统层面错误
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	return nil
}

func (svc *sessionService) UpdateUnread(ctx context.Context, userID int64, sessionID int64, updates model2.UpdateUnread) error {
	if err := svc.sessionRepo.UpdateUnread(ctx, userID, sessionID, updates); err != nil {
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}
	return nil
}

func (svc *sessionService) ClearUnread(ctx context.Context, userID int64, sessionID int64) error {
	if err := svc.sessionRepo.ClearUnread(ctx, userID, sessionID); err != nil {
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}
	return nil
}

func (svc *sessionService) CreateMessage(ctx context.Context, message *model2.Message) (*model2.Message, error) {
	messageModel := &model2.Message{
		ID:          svc.idGen.NextID(), // 补充 ID
		SessionID:   message.SessionID,
		SessionType: message.SessionType,
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		Content:     message.Content,
	}

	if err := svc.messageRepo.Create(ctx, messageModel); err != nil {
		slog.Error("Server Internal Error", "error", err)
		return &model2.Message{}, errs.ErrInternal
	}

	return messageModel, nil
}
