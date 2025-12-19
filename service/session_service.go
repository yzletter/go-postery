package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	sessiondto "github.com/yzletter/go-postery/dto/session"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
)

type sessionService struct {
	sessionRepo repository.SessionRepository
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	mqConn      *amqp.Connection
}

func NewSessionService(sessionRepo repository.SessionRepository, messageRepo repository.MessageRepository, userRepo repository.UserRepository, mq *amqp.Connection) SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
		mqConn:      mq,
	}
}

func (svc *sessionService) ListByUid(ctx context.Context, uid int64) ([]sessiondto.DTO, error) {
	var empty []sessiondto.DTO
	sessions, err := svc.sessionRepo.ListByUid(ctx, uid)
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	var sessionDTOs []sessiondto.DTO
	for _, session := range sessions {
		// 获取对方字段
		if session.TargetType == 1 {
			// 私聊
			targetUser, err := svc.userRepo.GetByID(ctx, session.TargetID)
			if err != nil {
				targetUser = &model.User{}
			}
			sessionDTO := sessiondto.ToDTO(session, targetUser)
			sessionDTOs = append(sessionDTOs, sessionDTO)
		} else {
			// todo 群聊查 Group 表
		}
	}

	return sessionDTOs, nil
}

func (svc *sessionService) Message(ctx context.Context, coon *websocket.Conn, uid, targetID int64) error {
	// todo
	panic(nil)
}

// Register 注册用户的 Exchange 和 Queue
func (svc *sessionService) Register(ctx context.Context, uid int64) error {
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
		_, err := ch.QueueDeclare(queueName, true, false, false, false, args)
		if err != nil {
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
			slog.Error("Queue Bind Failed", "queue_name", queueName)
			return errno.ErrServerInternal
		}
	}

	return nil
}
