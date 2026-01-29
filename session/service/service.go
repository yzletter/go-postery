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
	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/session/dto"
	"github.com/yzletter/go-postery/session/model"
	"github.com/yzletter/go-postery/session/repository"
	"github.com/yzletter/go-postery/session/service/ports"
	"github.com/yzletter/go-postery/session/utils"
)

type sessionService struct {
	sessionRepo   repository.SessionRepository
	messageRepo   repository.MessageRepository
	mqConn        *amqp.Connection
	kafkaConsumer *kafka.Reader
	idGen         ports.IDGenerator
	session_grpc.UnimplementedSessionServiceServer
}

func NewSessionService(sessionRepo repository.SessionRepository, messageRepo repository.MessageRepository, mq *amqp.Connection, consumer *kafka.Reader, idGen ports.IDGenerator) SessionService {
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
			var payload model.RegisterSessionEventPayload
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

func (svc *sessionService) ListByUID(ctx context.Context, id *session_grpc.UserID) (*session_grpc.Sessions, error) {
	var empty = new(session_grpc.Sessions)
	sessions, err := svc.sessionRepo.ListByUid(ctx, id.UserID)
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	var respSessions []*session_grpc.Session
	for _, session := range sessions {
		// 获取对方字段
		sessionDTO := dto.ToSession(session)
		respSessions = append(respSessions, sessionDTO)
	}

	return &session_grpc.Sessions{Sessions: respSessions}, nil
}

func (svc *sessionService) GetSession(ctx context.Context, id *session_grpc.BothUserID) (*session_grpc.Session, error) {
	var empty = new(session_grpc.Session)

	session, err := svc.sessionRepo.GetByUidAndTargetID(ctx, id.UserID, id.TargetID)
	if err != nil {
		// 系统层面错误
		if !errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrServerInternal
		}

		// 查找对方 session
		session, err := svc.sessionRepo.GetByUidAndTargetID(ctx, id.TargetID, id.UserID)
		if err != nil {
			// 系统层面错误
			if !errors.Is(err, repository.ErrRecordNotFound) {
				return empty, errno.ErrServerInternal
			}

			// 双边都没找到，新建会话
			ssid := svc.idGen.NextID()
			newSession1 := &model.Session{
				ID:         svc.idGen.NextID(),
				SessionID:  ssid,
				UserID:     id.UserID,
				TargetID:   id.TargetID,
				TargetType: 1,
			}

			newSession2 := &model.Session{
				ID:         svc.idGen.NextID(),
				SessionID:  ssid,
				UserID:     id.TargetID,
				TargetID:   id.UserID,
				TargetType: 1,
			}

			err = svc.sessionRepo.Create(ctx, newSession1)
			if err != nil {
				return empty, errno.ErrServerInternal
			}

			err = svc.sessionRepo.Create(ctx, newSession2)
			if err != nil {
				return empty, errno.ErrServerInternal
			}
			return dto.ToSession(newSession1), nil
		} else {
			// 对方的会话有，说明只有我单边删除，同一个 sessionID 单边新建
			ssid := session.SessionID
			newSession1 := &model.Session{
				ID:         svc.idGen.NextID(),
				SessionID:  ssid,
				UserID:     id.UserID,
				TargetID:   id.TargetID,
				TargetType: 1,
			}

			err = svc.sessionRepo.Create(ctx, newSession1)
			if err != nil {
				return empty, errno.ErrServerInternal
			}
			return dto.ToSession(newSession1), nil
		}
	}

	return dto.ToSession(session), nil
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
		return errno.ErrServerInternal
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
		return errno.ErrServerInternal
	}

	args := amqp.Table{
		"x-message-ttl":          int32(14 * 24 * 3600 * 1000), // 消息过期 TTL
		"x-dead-letter-exchange": "dlx",                        // 过期消息丢入死信队列
	}
	for _, queueName := range queueNames {
		// 申明队列
		if _, err := ch.QueueDeclare(queueName, true, false, false, false, args); err != nil {
			slog.Error("Queue Declare Failed", "uid", uid)
			return errno.ErrServerInternal
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
			return errno.ErrServerInternal
		}
	}

	return nil
}

func (svc *sessionService) GetHistoryMessagesByPage(ctx context.Context, req *session_grpc.GetHistoryMessagesByPageRequest) (*session_grpc.GetHistoryMessagesByPageResponse, error) {
	var empty = new(session_grpc.GetHistoryMessagesByPageResponse)
	total, messages, err := svc.messageRepo.GetByPage(ctx, req.UserID, req.TargetID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	var respMessages []*session_grpc.Message
	for _, message := range messages {
		respMessages = append(respMessages, dto.ToMessage(message))
	}

	return &session_grpc.GetHistoryMessagesByPageResponse{
		Count:    uint64(total),
		Messages: respMessages,
	}, nil
}

func (svc *sessionService) Delete(ctx context.Context, req *session_grpc.DeleteRequest) (*session_grpc.SessionEmptyResponse, error) {
	// 查当前用户这边的会话
	session, err := svc.sessionRepo.GetByID(ctx, req.UserID, req.SessionID)
	if err != nil {
		// 幂等
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &session_grpc.SessionEmptyResponse{}, nil
		}
		// 系统层面错误
		return &session_grpc.SessionEmptyResponse{}, errno.ErrServerInternal
	}

	if session.UserID != req.UserID {
		return &session_grpc.SessionEmptyResponse{}, errno.ErrUnauthorized
	}

	// 删除当前用户这边的会话, 要传 uid
	if err := svc.sessionRepo.Delete(ctx, req.UserID, req.SessionID); err != nil {
		// 幂等
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &session_grpc.SessionEmptyResponse{}, nil
		}
		// 系统层面错误
		return &session_grpc.SessionEmptyResponse{}, errno.ErrServerInternal
	}

	return &session_grpc.SessionEmptyResponse{}, nil
}

func (svc *sessionService) UpdateUnread(ctx context.Context, req *session_grpc.UpdateUnreadRequest) (*session_grpc.SessionEmptyResponse, error) {
	updates := model.UpdateUnread{
		Updates: model.Updates{
			LastMessageID:   req.LastMessageID,
			LastMessage:     req.LastMessage,
			LastMessageTime: utils.RPCTimeToGoTime(req.LastMessageTime),
		},
		Delta: int(req.Delta),
	}
	if err := svc.sessionRepo.UpdateUnread(ctx, req.UserID, req.SessionID, updates); err != nil {
		return &session_grpc.SessionEmptyResponse{}, errno.ErrServerInternal
	}
	return &session_grpc.SessionEmptyResponse{}, nil
}

func (svc *sessionService) ClearUnread(ctx context.Context, req *session_grpc.ClearUnreadRequest) (*session_grpc.SessionEmptyResponse, error) {
	if err := svc.sessionRepo.ClearUnread(ctx, req.UserID, req.SessionID); err != nil {
		return &session_grpc.SessionEmptyResponse{}, errno.ErrServerInternal
	}
	return &session_grpc.SessionEmptyResponse{}, nil
}

func (svc *sessionService) CreateMessage(ctx context.Context, message *session_grpc.Message) (*session_grpc.Message, error) {
	messageModel := &model.Message{
		ID:          svc.idGen.NextID(), // 补充 ID
		SessionID:   message.SessionID,
		SessionType: int(message.SessionType),
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		Content:     message.Content,
	}

	if err := svc.messageRepo.Create(ctx, messageModel); err != nil {
		return &session_grpc.Message{}, errno.ErrServerInternal
	}

	return &session_grpc.Message{
		ID:          messageModel.ID,
		SessionID:   messageModel.SessionID,
		SessionType: int32(messageModel.SessionType),
		MessageFrom: messageModel.MessageFrom,
		MessageTo:   messageModel.MessageTo,
		Content:     messageModel.Content,
		CreatedAt:   utils.GoTimeToRPCTime(&messageModel.CreatedAt),
	}, nil
}
