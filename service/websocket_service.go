package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	messagedto "github.com/yzletter/go-postery/dto/message"
	sessiondto "github.com/yzletter/go-postery/dto/session"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
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

type websocketService struct {
	sessionRepo repository.SessionRepository
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	mqConn      *amqp.Connection
	idGen       ports.IDGenerator
}

func NewWebsocketService(sessionRepo repository.SessionRepository, messageRepo repository.MessageRepository, userRepo repository.UserRepository,
	mq *amqp.Connection, idGen ports.IDGenerator) WebsocketService {
	return &websocketService{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
		mqConn:      mq,
		idGen:       idGen,
	}
}

func (svc *websocketService) Connect(ctx context.Context, w http.ResponseWriter, r *http.Request, uid int64) error {
	// 升级 HTTP 连接
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// 升级失败
		return errno.ErrServerInternal
	}
	defer wsConn.Close()

	// 心跳保持
	go heartBeat(wsConn)

	buffer := make(chan model.Message, 100)

	// 子程：写数据到 Websocket 中
	go consumeMQ(ctx, svc.mqConn, wsConn, uid, buffer)

	// 主程：从 Websocket 中读数据
	for {
		// 从 Websocket 中读取消息
		_, messageBody, err := wsConn.ReadMessage() // 如果对方主动断开连接或超时，该行会报错，for 循环会退出
		if err != nil {
			break
		}

		var message model.Message
		err = json.Unmarshal(messageBody, &message)
		if err != nil {
			// BadRequest
			continue
		}

		// 过滤消息
		ok := intercept(message)
		if !ok {
			continue
		}

		// 落库
		message.ID = svc.idGen.NextID() // 	补全 ID

		err = svc.messageRepo.Create(ctx, &message)
		if err != nil {
			slog.Error("Connect Store Failed", "message", message)
			continue
		}

		// 更新对方的会话信息
		contentBrief := []rune(message.Content) // 最后一条消息的摘要
		if len(contentBrief) > 5 {
			contentBrief = contentBrief[:5]
		}
		updates := sessiondto.UpdateUnreadRequest{
			LastMessageID:   message.ID,
			LastMessage:     string(contentBrief),
			LastMessageTime: message.CreatedAt,
		}
		err = svc.sessionRepo.UpdateUnread(ctx, message.MessageTo, message.SessionID, updates)
		if err != nil {
			slog.Error("Update Unread Failed", "user_id", message.MessageTo, "error", err)
		}

		// 发给 MQ
		err = produceMQ(ctx, svc.mqConn, message, message.MessageTo)
		if err != nil {
			slog.Error("Produce To MQ Failed", "id", message.MessageTo, "error", err)
		}
		err = produceMQ(ctx, svc.mqConn, message, message.MessageFrom)
		if err != nil {
			slog.Error("Produce To MQ Failed", "id", message.MessageFrom, "error", err)
		}
	}

	return nil
}

// 心跳保持
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

// 消费队列
func consumeMQ(ctx context.Context, mqConn *amqp.Connection, wsConn *websocket.Conn, id int64, buffer chan model.Message) error {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Receive Failed", "error", err)
		}
	}()

	// 消费的队列名
	queueName := fmt.Sprintf("%d_computer", id)

	ch, err := mqConn.Channel()
	if err != nil {
		return err
	}

	// 开始消费队列写入 Buffer
	deliverCh, _ := ch.ConsumeWithContext(ctx, queueName, "", false, false, false, false, nil)
	go func() {
		for deliver := range deliverCh {
			var message model.Message
			_ = json.Unmarshal(deliver.Body, &message)

			buffer <- message
			deliver.Ack(false) // ACK
		}
	}()

	// 从 Buffer 中读取数据打给前端
	go func() {
		for {
			msg := <-buffer
			msgDTO := messagedto.ToDTO(&msg)
			wsConn.WriteJSON(msgDTO) // 打给前端
		}
	}()

	return nil
}

// 将消息发给 MQ 的 Exchange
func produceMQ(ctx context.Context, conn *amqp.Connection, message model.Message, id int64) error {
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
