package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/kafka-go"
	ws_gateway_grpc "github.com/yzletter/go-postery/api/proto/ws_gateway/v1"
	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/session/domain"
	session_bff_dto "github.com/yzletter/go-postery/backend/micro/session/dto"
	repository "github.com/yzletter/go-postery/backend/micro/session/repository"
	"github.com/yzletter/go-postery/backend/ports"
)

type sessionService struct {
	wsGateway     manager.WSGatewayClient
	sessionRepo   repository.SessionRepository
	messageRepo   repository.MessageRepository
	mqConn        *amqp.Connection
	kafkaConsumer *kafka.Reader
	idGen         ports.IDGenerator
}

func NewSessionService(wsGateway manager.WSGatewayClient, sessionRepo repository.SessionRepository, messageRepo repository.MessageRepository, mq *amqp.Connection, consumer *kafka.Reader, idGen ports.IDGenerator) SessionService {
	return &sessionService{
		wsGateway:     wsGateway,
		sessionRepo:   sessionRepo,
		messageRepo:   messageRepo,
		mqConn:        mq,
		kafkaConsumer: consumer,
		idGen:         idGen,
	}
}

// 消费队列
func (svc *sessionService) consumeMQ(ctx context.Context, mqConn *amqp.Connection, id int64) (retErr error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Receive Failed", "error", err)
			retErr = fmt.Errorf("consume session queue panic: %v", err)
		}
	}()
	if svc.wsGateway == nil {
		return errs.ErrUnavailable
	}

	// 消费的队列名
	queueName := fmt.Sprintf("%d_computer", id)

	ch, err := mqConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 开始消费队列并写入 Websocket
	deliverCh, err := ch.ConsumeWithContext(ctx, queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case deliver, ok := <-deliverCh:
			if !ok {
				return nil
			}
			var message domain.Message

			if err := json.Unmarshal(deliver.Body, &message); err != nil {
				slog.Error("Unmarshal MQ message failed", "error", err)
				_ = deliver.Nack(false, false)
				continue
			}
			msgDTO := session_bff_dto.ToMessageDTO(&message)
			data, err := sonic.Marshal(msgDTO)
			if err != nil {
				_ = deliver.Nack(false, false)
				return err
			}
			if _, err := svc.wsGateway.Push(ctx, &ws_gateway_grpc.PushRequest{
				UserID:  id,
				BizType: "session",
				BizData: data,
			}); err != nil {
				return err
			}
			if err := deliver.Ack(false); err != nil {
				slog.Error("ACK MQ message failed", "error", err)
			}
		}
	}
}

func (svc *sessionService) NewConnection(ctx context.Context, uid int64) error {
	// WebSocket 断开时 ctx 会取消，consumeMQ 随之停止消费并关闭 channel。
	if err := svc.consumeMQ(ctx, svc.mqConn, uid); err != nil {
		if ctx.Err() == nil {
			slog.Error("consume session queue failed", "user_id", uid, "error", err)
		}
		return err
	}

	return nil
}

func (svc *sessionService) Chat(ctx context.Context, uid int64, message domain.Message) error {
	// 过滤消息
	ok := intercept(message, uid)
	if !ok {
		slog.Warn("chat rejected: sender mismatch", "user_id", uid, "message_from", message.MessageFrom)
		return errs.ErrUnauthenticated
	}

	session, err := svc.GetSession(ctx, uid, message.MessageTo)
	if err != nil {
		return err
	}
	if session.SessionID != message.SessionID {
		slog.Warn("chat rejected: session mismatch", "user_id", uid, "session_id", message.SessionID, "actual_session_id", session.SessionID)
		return errs.ErrInvalidArgument
	}

	// 消息落库
	message, err = svc.CreateMessage(ctx, message)
	if err != nil {
		return err
	}

	// 更新会话信息
	contentBrief := []rune(message.Content) // 最后一条消息的摘要
	if len(contentBrief) > 5 {
		contentBrief = contentBrief[:5]
	}

	// 更新对方会话信息, 增加未读
	if err := svc.UpdateUnread(ctx, message.MessageTo, message.SessionID, domain.UpdateUnread{
		Updates: domain.Updates{
			LastMessageID:   message.ID,
			LastMessage:     string(contentBrief),
			LastMessageTime: message.CreatedAt,
		},
		Delta: 1,
	}); err != nil {
		slog.Error("Update Unread Failed", "user_id", message.MessageTo, "error", err)
	}

	// 更新己方会话信息，不增加未读数。
	if err := svc.UpdateUnread(ctx, message.MessageFrom, message.SessionID, domain.UpdateUnread{
		Updates: domain.Updates{
			LastMessageID:   message.ID,
			LastMessage:     string(contentBrief),
			LastMessageTime: message.CreatedAt,
		},
		Delta: 0,
	}); err != nil {
		slog.Error("Update Unread Failed", "user_id", message.MessageFrom, "error", err)
	}

	// 双向投递, 发给 MQ
	if err = produceMQ(ctx, svc.mqConn, message, message.MessageTo); err != nil {
		slog.Error("Produce To MQ Failed", "id", message.MessageTo, "error", err)
	}

	if err = produceMQ(ctx, svc.mqConn, message, message.MessageFrom); err != nil {
		slog.Error("Produce To MQ Failed", "id", message.MessageFrom, "error", err)
	}

	return nil
}

// 将消息发给 MQ 的 Exchange
func produceMQ(ctx context.Context, conn *amqp.Connection, message domain.Message, id int64) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msg, _ := json.Marshal(message)

	exchangeName := fmt.Sprintf("%d_exchange", id)
	err = ch.PublishWithContext(
		ctx,
		exchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json", // MIME content type
			Body:         msg,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// todo 处理消息内容, 正常应进行对非法内容进行拦截。比如机器人消息（发言频率过快）；包含欺诈、涉政等违规内容；涉嫌私下联系/交易等。
func intercept(message domain.Message, uid int64) bool {
	if message.MessageFrom != uid {
		return false
	}
	return true
}

func (svc *sessionService) StartSessionRegisterConsumer(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			slog.Info("session register consumer closed")
			return
		default:
			message, err := svc.kafkaConsumer.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { // 正常退出
					return
				}

				slog.Warn("fetch session register event failed", "backoff", backoff, "error", err)

				// 简单退避，避免狂刷日志
				time.Sleep(backoff)
				if backoff < 10*time.Second {
					backoff *= 2
				}
				continue
			}

			backoff = time.Second
			var payload event.NewUserEventPayload
			if err := sonic.Unmarshal(message.Value, &payload); err != nil {
				// 脏消息
				slog.Warn("invalid session register event, skip", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "error", err)
				_ = svc.kafkaConsumer.CommitMessages(ctx, message) // 把脏消息 Commit 掉，避免卡住
				continue
			}

			slog.Info("session register event received", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "uid", payload.ID)

			// 进行注册, 幂等
			if err := svc.register(ctx, payload.ID); err != nil {
				slog.Error("register session queue failed", "uid", payload.ID, "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "error", err)
				time.Sleep(time.Second) // 最小退避，避免打爆
				continue                // 不 commit -> 重试
			}

			// 把消息 Commit 掉
			if err := svc.kafkaConsumer.CommitMessages(ctx, message); err != nil {
				slog.Error("commit session register event failed", "uid", payload.ID, "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "error", err)

				// Commit 失败通常会导致重复消费，但不会丢消息，可接受
				continue
			}
		}
	}
}

func (svc *sessionService) ListByUID(ctx context.Context, userID int64) ([]domain.Session, error) {
	sessions, err := svc.sessionRepo.ListByUid(ctx, userID)
	if err != nil {
		slog.Error("list sessions failed", "user_id", userID, "error", err)
		return nil, errs.ErrInternal
	}

	return sessions, nil
}

func (svc *sessionService) GetSession(ctx context.Context, userID int64, targetID int64) (domain.Session, error) {
	session, err := svc.sessionRepo.GetByUidAndTargetID(ctx, userID, targetID)
	if err != nil {
		// 系统层面错误
		if !errors.Is(err, repository.ErrRecordNotFound) {
			slog.Error("get session failed", "user_id", userID, "target_id", targetID, "error", err)
			return domain.Session{}, errs.ErrInternal
		}

		// 查找对方 session
		session, err := svc.sessionRepo.GetByUidAndTargetID(ctx, targetID, userID)
		if err != nil {
			// 系统层面错误
			if !errors.Is(err, repository.ErrRecordNotFound) {
				slog.Error("get peer session failed", "user_id", targetID, "target_id", userID, "error", err)
				return domain.Session{}, errs.ErrInternal
			}

			// 双边都没找到，新建会话
			ssid := svc.idGen.NextID()
			newSession1 := domain.Session{
				ID:         svc.idGen.NextID(),
				SessionID:  ssid,
				UserID:     userID,
				TargetID:   targetID,
				TargetType: 1,
			}

			newSession2 := domain.Session{
				ID:         svc.idGen.NextID(),
				SessionID:  ssid,
				UserID:     targetID,
				TargetID:   userID,
				TargetType: 1,
			}

			if err = svc.sessionRepo.Create(ctx, newSession1, newSession2); err != nil {
				slog.Error("create session failed", "user_id", userID, "target_id", targetID, "session_id", ssid, "error", err)
				return domain.Session{}, errs.ErrInternal
			}

			return newSession1, nil
		}

		// 对方的会话有，说明只有我单边删除，同一个 sessionID 单边恢复
		ssid := session.SessionID
		if session, err := svc.sessionRepo.Recover(ctx, userID, targetID); err == nil {
			return session, nil
		}
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("recover session skipped: deleted session not found", "user_id", userID, "target_id", targetID, "session_id", ssid)
		} else {
			slog.Warn("recover session failed", "user_id", userID, "target_id", targetID, "session_id", ssid, "error", err)
		}

		// 恢复失败进行创建
		newSession1 := domain.Session{
			ID:         svc.idGen.NextID(),
			SessionID:  ssid,
			UserID:     userID,
			TargetID:   targetID,
			TargetType: 1,
		}

		err = svc.sessionRepo.Create(ctx, newSession1)
		if err != nil {
			slog.Error("create one-side session failed", "user_id", userID, "target_id", targetID, "session_id", ssid, "error", err)
			return domain.Session{}, errs.ErrInternal
		}
		return newSession1, nil
	}

	return session, nil
}

// register 注册用户的 Exchange 和 Queue
func (svc *sessionService) register(ctx context.Context, uid int64) error {

	// 定义 Exchange 和 Queue 名称
	exchangeName := fmt.Sprintf("%d_exchange", uid)
	queueNameComputer := fmt.Sprintf("%d_computer", uid)
	queueNameMobile := fmt.Sprintf("%d_mobile", uid)
	queueNames := []string{queueNameComputer, queueNameMobile}

	ch, err := svc.mqConn.Channel()
	if err != nil {
		slog.Error("open rabbitmq channel failed", "uid", uid, "error", err)
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
		slog.Error("declare session exchange failed", "uid", uid, "error", err)
		return errs.ErrInternal
	}

	args := amqp.Table{
		"x-message-ttl":          int32(14 * 24 * 3600 * 1000), // 消息过期 TTL
		"x-dead-letter-exchange": "dlx",                        // 过期消息丢入死信队列
	}
	for _, queueName := range queueNames {
		// 声明队列
		if _, err := ch.QueueDeclare(queueName, true, false, false, false, args); err != nil {
			slog.Error("declare session queue failed", "uid", uid, "queue", queueName, "error", err)
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
			slog.Error("bind session queue failed", "queue", queueName, "exchange", exchangeName, "error", err)
			return errs.ErrInternal
		}
	}

	return nil
}

func (svc *sessionService) GetHistoryMessagesByPage(ctx context.Context, userID int64, targetID int64, pageNo int, pageSize int) (int64, []domain.Message, error) {
	total, messages, err := svc.messageRepo.GetByPage(ctx, userID, targetID, pageNo, pageSize)
	if err != nil {
		slog.Error("get history messages failed", "user_id", userID, "target_id", targetID, "page_no", pageNo, "page_size", pageSize, "error", err)
		return 0, nil, errs.ErrInternal
	}

	return int64(total), messages, nil
}

func (svc *sessionService) Delete(ctx context.Context, userID int64, sessionID int64) error {
	// 查当前用户侧会话
	session, err := svc.sessionRepo.GetByID(ctx, userID, sessionID)
	if err != nil {
		// 幂等
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("delete session skipped: session not found", "user_id", userID, "session_id", sessionID)
			return nil
		}
		// 系统层面错误
		slog.Error("get session before delete failed", "user_id", userID, "session_id", sessionID, "error", err)
		return errs.ErrInternal
	}

	if session.UserID != userID {
		slog.Info("delete session rejected: unauthenticated", "user_id", userID, "session_id", sessionID)
		return errs.ErrUnauthenticated
	}

	// 软删除当前用户侧会话
	if err := svc.sessionRepo.Delete(ctx, userID, sessionID); err != nil {
		// 幂等
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("delete session skipped: session not found", "user_id", userID, "session_id", sessionID)
			return nil
		}
		// 系统层面错误
		slog.Error("delete session failed", "user_id", userID, "session_id", sessionID, "error", err)
		return errs.ErrInternal
	}

	return nil
}

func (svc *sessionService) UpdateUnread(ctx context.Context, userID int64, sessionID int64, updates domain.UpdateUnread) error {
	if err := svc.sessionRepo.UpdateUnread(ctx, userID, sessionID, updates); err != nil {
		slog.Error("update unread failed", "user_id", userID, "session_id", sessionID, "error", err)
		return errs.ErrInternal
	}
	return nil
}

func (svc *sessionService) ClearUnread(ctx context.Context, userID int64, sessionID int64) error {
	if err := svc.sessionRepo.ClearUnread(ctx, userID, sessionID); err != nil {
		slog.Error("clear unread failed", "user_id", userID, "session_id", sessionID, "error", err)
		return errs.ErrInternal
	}
	return nil
}

func (svc *sessionService) CreateMessage(ctx context.Context, message domain.Message) (domain.Message, error) {
	now := time.Now()
	messageModel := domain.Message{
		ID:          svc.idGen.NextID(), // 补充 ID
		SessionID:   message.SessionID,
		SessionType: message.SessionType,
		MessageFrom: message.MessageFrom,
		MessageTo:   message.MessageTo,
		Content:     message.Content,
		// Repository 接口按值传递，需在这里补齐时间供会话更新和 WebSocket 下行使用。
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := svc.messageRepo.Create(ctx, messageModel); err != nil {
		slog.Error("create message failed", "session_id", message.SessionID, "from", message.MessageFrom, "to", message.MessageTo, "error", err)
		return domain.Message{}, errs.ErrInternal
	}

	return messageModel, nil
}
