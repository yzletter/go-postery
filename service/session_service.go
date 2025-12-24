package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	messagedto "github.com/yzletter/go-postery/dto/message"
	sessiondto "github.com/yzletter/go-postery/dto/session"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
	"github.com/yzletter/go-postery/utils/response"
)

var (
	pongWait   = 5 * time.Second // 等待 pong 的超时时间
	pingPeriod = 3 * time.Second // 发送 ping 的周期，必须短于 pongWait
)

// HTTP 升级器
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	ReadBufferSize:   10000,
	WriteBufferSize:  10000,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

		// 没找到，新建会话
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

func (svc *sessionService) Message(ctx *gin.Context, uid, targetID int64) error {
	// 升级 HTTP
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		slog.Error("Upgrade HTTP Failed", "error", err)
		response.Error(ctx, errno.ErrServerInternal)
		return err
	}

	// 心跳保持
	go heartBeat(conn)

	defer conn.Close()

	buffer := make(chan model.Message, 100)

	// 消费 MQ 中消息写给 buffer
	go func() {
		err := consume(ctx, svc.mqConn, uid, targetID, buffer)
		if err != nil {
			slog.Error("Consume Failed", "error", err)
		}
	}()

	// 从 buffer 中读取消息写给 WS
	go func() {
		err := svc.receive(ctx, conn, uid, targetID, buffer)
		if err != nil {
			slog.Error("Consume Failed", "error", err)
		}
	}()

	// 从 WS 中读取消息写给 MQ
	for {
		// 从 Websocket 中读取消息
		_, body, err := conn.ReadMessage() // 如果对方主动断开连接或超时，该行会报错，for 循环会退出
		if err != nil {
			break
		}

		var message model.Message
		err = json.Unmarshal(body, &message)
		if err != nil {
			// BadRequest
			continue
		}

		// 过滤
		ok := intercept(message)
		if !ok {
			continue
		}

		// 落库
		message.ID = svc.idGen.NextID() // 	补全 ID

		err = svc.messageRepo.Create(ctx, &message)
		if err != nil {
			slog.Error("Message Store Failed", "message", message)
			continue
		}

		// 发给 MQ
		err = produce(ctx, svc.mqConn, message, message.MessageTo)
		if err != nil {
			slog.Error("Produce To MQ Failed", "id", message.MessageTo, "error", err)
		}

		err = produce(ctx, svc.mqConn, message, message.MessageFrom)
		if err != nil {
			slog.Error("Produce To MQ Failed", "id", message.MessageFrom, "error", err)
		}
	}

	return nil
}

func (svc *sessionService) receive(ctx context.Context, conn *websocket.Conn, id int64, targetID int64, buffer chan model.Message) error {
	defer func() {
		err := conn.Close() // 错误忽略
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	// 加载历史消息
	msgs, err := svc.messageRepo.GetByIDAndTargetID(ctx, id, targetID)
	if err != nil {
		return errno.ErrServerInternal
	}

	// 哈希表
	Set := make(map[messagedto.DTO]struct{})
	for _, msg := range msgs {
		msgDTO := messagedto.ToDTO(msg)
		Set[msgDTO] = struct{}{}
		conn.WriteJSON(msgDTO) // 打给前端
	}

	// 从 buffer 中读取数据
	for {
		msg := <-buffer
		msgDTO := messagedto.ToDTO(&msg)
		// 判断是否加载过
		if _, exits := Set[msgDTO]; !exits {
			Set[msgDTO] = struct{}{}
			conn.WriteJSON(msgDTO) // 打给前端
		}
	}
}

// 消费对应 MQ 的 Queue
func consume(ctx context.Context, conn *amqp.Connection, id int64, targetID int64, buffer chan model.Message) error {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Receive Failed", "error", err)
		}
	}()

	// 消费的队列名
	queueName := fmt.Sprintf("%d_computer", id)

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	// 开始消费队列
	deliverCh, _ := ch.ConsumeWithContext(ctx, queueName, "", false, false, false, false, nil)
	go func() {
		for deliver := range deliverCh {
			var message model.Message
			_ = json.Unmarshal(deliver.Body, &message)

			if targetID == message.MessageFrom || targetID == message.MessageTo {
				buffer <- message
				deliver.Ack(false) // ACK
			}
		}
	}()

	return nil
}

// 将消息发给 MQ 的 Exchange
func produce(ctx context.Context, conn *amqp.Connection, message model.Message, id int64) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

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

func heartBeat(conn *websocket.Conn) {
	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	err := conn.WriteMessage(websocket.PingMessage, nil)
	if err != nil {
		conn.WriteMessage(websocket.CloseMessage, nil)
	}

	ticker := time.NewTicker(pingPeriod)
LOOP:
	for {
		<-ticker.C
		err := conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			conn.WriteMessage(websocket.CloseMessage, nil)
			break LOOP
		}
		deadline := time.Now().Add(pongWait) // ping发出去以后，期望5秒之内从conn里能计到数据（至少能读到pong）
		conn.SetReadDeadline(deadline)
	}
}

// todo 处理消息内容, 正常应进行对非法内容进行拦截。比如机器人消息（发言频率过快）；包含欺诈、涉政等违规内容；涉嫌私下联系/交易等。
func intercept(message model.Message) bool {
	return true
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
