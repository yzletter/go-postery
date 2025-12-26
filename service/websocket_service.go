package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	// 校验 Origin
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// 没有 Origin（比如用 Postman、命令行测试），直接放行
			return true
		}

		// 允许的前端地址白名单
		allowList := map[string]bool{
			"http://localhost:5173": true,
		}

		if allowList[origin] {
			return true
		}

		// 同源也放行
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return strings.EqualFold(u.Host, r.Host)
	},
}

type wsWriteRequest struct {
	messageType int
	data        []byte
	jsonPayload interface{}
	isJSON      bool
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

	connCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		fmt.Println("close websocket connection")
		wsConn.Close()
	}()

	writeCh := make(chan wsWriteRequest, 50)
	send := func(req wsWriteRequest) bool {
		select {
		case <-connCtx.Done():
			return false
		case writeCh <- req:
			return true
		}
	}

	// 单独 writer 协程，避免并发写 conn
	go func() {
		defer cancel()
		for {
			select {
			case <-connCtx.Done():
				return
			case req, ok := <-writeCh:
				if !ok {
					return
				}
				var err error
				// 判断传入的 req
				if req.isJSON { // 是 JSON
					err = wsConn.WriteJSON(req.jsonPayload)
				} else { // 是其他
					err = wsConn.WriteMessage(req.messageType, req.data)
				}
				if err != nil {
					slog.Error("Websocket write failed", "error", err)
					return
				}
			}
		}
	}()

	// 心跳保持
	go heartBeat(connCtx, wsConn, send)

	// 子程：写数据到 Websocket 中
	go func() {
		if err := consumeMQ(connCtx, svc.mqConn, uid, send); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("Consume MQ Failed", "error", err)
		}
	}()

	// 主程：从 Websocket 中读数据
	for {
		// 从 Websocket 中读取消息
		_, messageBody, err := wsConn.ReadMessage() // 如果对方主动断开连接或超时，该行会报错，for 循环会退出
		if err != nil {
			break
		}

		var messageReq messagedto.Request
		err = json.Unmarshal(messageBody, &messageReq)
		if err != nil {
			// BadRequest
			continue
		}

		if messageReq.Type == "read_ack" {
			// 是 read_ack
			ssid, err := strconv.ParseInt(messageReq.SessionID, 10, 64)
			if err != nil {
				continue
			}
			svc.sessionRepo.ClearUnread(ctx, uid, ssid)
		} else if messageReq.Type == "message" {
			// 是消息
			ssid, err := strconv.ParseInt(messageReq.SessionID, 10, 64)
			if err != nil {
				continue
			}

			messageFrom, err := strconv.ParseInt(messageReq.MessageFrom, 10, 64)
			if err != nil {
				continue
			}

			messageTo, err := strconv.ParseInt(messageReq.MessageTo, 10, 64)
			if err != nil {
				continue
			}

			message := model.Message{
				ID:          svc.idGen.NextID(),
				SessionID:   ssid,
				SessionType: messageReq.SessionType,
				MessageFrom: messageFrom,
				MessageTo:   messageTo,
				Content:     messageReq.Content,
			}

			// 过滤消息
			ok := intercept(message, uid)
			if !ok {
				continue
			}

			// 落库
			message.ID = svc.idGen.NextID()                                          // 补全 ID
			session, err := svc.sessionRepo.GetByUidAndTargetID(ctx, uid, messageTo) // 查找 session
			if err != nil || session.SessionID != ssid {
				continue
			}

			err = svc.messageRepo.Create(ctx, &message)
			if err != nil {
				slog.Error("Connect Store Failed", "message", message)
				continue
			}

			// 更新会话信息
			contentBrief := []rune(message.Content) // 最后一条消息的摘要
			if len(contentBrief) > 5 {
				contentBrief = contentBrief[:5]
			}
			updates := sessiondto.Updates{
				LastMessageID:   message.ID,
				LastMessage:     string(contentBrief),
				LastMessageTime: message.CreatedAt,
			}

			// 更新对方会话信息, 增加未读
			err = svc.sessionRepo.UpdateUnread(ctx, message.MessageTo, message.SessionID, sessiondto.UpdateUnreadRequest{Updates: updates, Delta: 1})
			if err != nil {
				slog.Error("Update Unread Failed", "user_id", message.MessageTo, "error", err)
			}
			// 更新己方会话信息, 不增加未读
			err = svc.sessionRepo.UpdateUnread(ctx, message.MessageFrom, message.SessionID, sessiondto.UpdateUnreadRequest{Updates: updates, Delta: 0})
			if err != nil {
				slog.Error("Update Unread Failed", "user_id", message.MessageFrom, "error", err)
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
	}

	cancel()
	return nil
}

// 心跳保持
func heartBeat(ctx context.Context, conn *websocket.Conn, send func(wsWriteRequest) bool) {
	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	// 启动 Ticker
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	if !send(wsWriteRequest{messageType: websocket.PingMessage}) {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !send(wsWriteRequest{messageType: websocket.PingMessage}) {
				return
			}
			deadline := time.Now().Add(pongWait) // ping发出去以后，期望5秒之内从conn里能计到数据（至少能读到pong）
			conn.SetReadDeadline(deadline)
			//fmt.Printf("must read before %s\n", deadline.Format("2006-01-02 15:04:05"))
		}
	}
}

// todo 处理消息内容, 正常应进行对非法内容进行拦截。比如机器人消息（发言频率过快）；包含欺诈、涉政等违规内容；涉嫌私下联系/交易等。
func intercept(message model.Message, uid int64) bool {
	if message.MessageFrom != uid {
		return false
	}
	return true
}

// 消费队列
func consumeMQ(ctx context.Context, mqConn *amqp.Connection, id int64, send func(wsWriteRequest) bool) error {
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
			var message model.Message
			if err := json.Unmarshal(deliver.Body, &message); err != nil {
				slog.Error("Unmarshal MQ message failed", "error", err)
				_ = deliver.Nack(false, false)
				continue
			}

			msgDTO := messagedto.ToDTO(&message)
			if !send(wsWriteRequest{isJSON: true, jsonPayload: msgDTO}) {
				return ctx.Err()
			}
			if err := deliver.Ack(false); err != nil {
				slog.Error("ACK MQ message failed", "error", err)
			}
		}
	}
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
