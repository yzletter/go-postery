package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	messagedto "github.com/yzletter/go-postery/dto/message"
	sessiondto "github.com/yzletter/go-postery/dto/session"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

var (
	pongWait   = 5 * time.Second // 等待 pong 的超时时间
	pingPeriod = 3 * time.Second // 发送 ping 的周期，必须短于 pongWait
)

type sessionService struct {
	sessionRepo repository.SessionRepository
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	mqConn      *amqp.Connection
	idGen       ports.IDGenerator
}

func (svc *sessionService) GetSession(ctx context.Context, uid, targetID int64) (sessiondto.DTO, error) {
	var empty sessiondto.DTO
	user, err := svc.userRepo.GetByID(ctx, targetID)
	if err != nil {
		user = &model.User{}
	}

	session, err := svc.sessionRepo.GetByUidAndTargetID(ctx, uid, targetID)
	if err != nil {
		// 系统层面错误
		if !errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrServerInternal
		}

		// 查找对方 session
		session, err := svc.sessionRepo.GetByUidAndTargetID(ctx, targetID, uid)
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
				UserID:     uid,
				TargetID:   targetID,
				TargetType: 1,
			}

			newSession2 := &model.Session{
				ID:         svc.idGen.NextID(),
				SessionID:  ssid,
				UserID:     targetID,
				TargetID:   uid,
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
			return sessiondto.ToDTO(newSession1, user), nil
		} else {
			// 对方的会话有，说明只有我单边删除，同一个 sessionID 单边新建
			ssid := session.SessionID
			newSession1 := &model.Session{
				ID:         svc.idGen.NextID(),
				SessionID:  ssid,
				UserID:     uid,
				TargetID:   targetID,
				TargetType: 1,
			}

			err = svc.sessionRepo.Create(ctx, newSession1)
			if err != nil {
				return empty, errno.ErrServerInternal
			}
			return sessiondto.ToDTO(newSession1, user), nil
		}
	}

	return sessiondto.ToDTO(session, user), nil
}

func NewSessionService(sessionRepo repository.SessionRepository, messageRepo repository.MessageRepository, userRepo repository.UserRepository,
	mq *amqp.Connection, idGen ports.IDGenerator) SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
		mqConn:      mq,
		idGen:       idGen,
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

func (svc *sessionService) GetHistoryMessagesByPage(ctx context.Context, uid int64, targetID int64, pageNo, pageSize int) (int, []messagedto.DTO, error) {
	var empty []messagedto.DTO
	total, messages, err := svc.messageRepo.GetByPage(ctx, uid, targetID, pageNo, pageSize)
	if err != nil {
		return 0, empty, errno.ErrServerInternal
	}

	var messageDTOs []messagedto.DTO
	for _, message := range messages {
		messageDTOs = append(messageDTOs, messagedto.ToDTO(message))
	}

	return total, messageDTOs, nil
}

func (svc *sessionService) Delete(ctx context.Context, uid, sid int64) error {
	// 查当前用户这边的会话
	session, err := svc.sessionRepo.GetByID(ctx, uid, sid)
	if err != nil {
		// 幂等
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil
		}
		// 系统层面错误
		return errno.ErrServerInternal
	}

	if session.UserID != uid {
		return errno.ErrUnauthorized
	}

	// 删除当前用户这边的会话, 要传 uid
	err = svc.sessionRepo.Delete(ctx, uid, sid)
	if err != nil {
		// 幂等
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil
		}
		// 系统层面错误
		return errno.ErrServerInternal
	}

	return nil
}
