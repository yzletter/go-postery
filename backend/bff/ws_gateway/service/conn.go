package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
	"github.com/yzletter/go-postery/backend/bff/ws_gateway/domain"
)

const (
	pongWait   = 5 * time.Second // pongWait 表示服务端等待客户端 pong 的最大时间
	pingPeriod = 3 * time.Second // pingPeriod 表示服务端发送 ping 的周期
)

// writeReq 表示一次 WebSocket 写请求
type writeReq struct {
	messageType int // PingMessage TextMessage 等
	data        []byte
}

// MyWebsocketConn 表示一个用户的一条 WebSocket 连接
type MyWebsocketConn struct {
	biz        domain.ConnType   // 连接类型
	userID     int64             // 所属用户
	gate       *WebsocketGateway // 所属网关
	conn       *websocket.Conn   // 底层连接
	writeCh    chan writeReq     // 写队列
	connCtx    context.Context   // 当前连接的上下文
	connCancel context.CancelFunc
	closeOnce  sync.Once
}

// Writer 写程
func (conn *MyWebsocketConn) Writer() {
	// Writer 退出时，说明写链路已经不可用, 从网关连接池中删除当前连接
	defer func() {
		_ = conn.gate.DelConnection(context.Background(), conn.userID, conn.biz, conn)
	}()

	for {
		select {
		case <-conn.connCtx.Done():
			return
		case req := <-conn.writeCh:
			// 设置写超时，避免 WriteMessage 永久阻塞
			_ = conn.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

			// 串行写 WebSocket
			if err := conn.conn.WriteMessage(req.messageType, req.data); err != nil {
				slog.Error("Websocket write failed", "userID", conn.userID, "error", err)
				return
			}
		}
	}
}

// Reader 是当前连接的读协程
func (conn *MyWebsocketConn) Reader() {
	// Reader 退出时，删除连接
	defer func() {
		_ = conn.gate.DelConnection(context.Background(), conn.userID, conn.biz, conn)
	}()

	_ = conn.conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.conn.SetReadLimit(1000000) // 设置上限, 防止大消息 OOM
	conn.conn.SetPongHandler(func(appData string) error {
		return conn.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		// 读取客户端消息
		_, messageBody, err := conn.conn.ReadMessage()
		if err != nil {
			slog.Info("Websocket read stopped", "userID", conn.userID, "error", err)
			return
		}

		var message WSMessage

		// 反序列化标准 WebSocket 信封。
		if err := sonic.Unmarshal(messageBody, &message); err != nil {
			slog.Warn("Invalid websocket message", "userID", conn.userID, "error", err)
			continue
		}

		// 如果没有配置 handler，就只读取，不做业务处理
		if conn.gate.handler == nil {
			continue
		}

		// 把消息交给业务 handler, 这里使用 c.connCtx，表示只要连接关闭，业务处理也应该感知到取消。
		if err := conn.gate.handler.HandleWSMessage(conn.connCtx, conn.userID, conn.biz, message); err != nil {
			slog.Error("Handle websocket message failed", "userID", conn.userID, "bizType", message.BizType, "error", err)
			continue
		}
	}
}

// HeartBeat 是心跳协程
func (conn *MyWebsocketConn) HeartBeat() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-conn.connCtx.Done():
			return
		case <-ticker.C:
			// 发送心跳
			if ok := conn.Send(context.Background(), writeReq{messageType: websocket.PingMessage}); !ok {
				return
			}
		}
	}
}

// Send 把消息写入当前连接的写队列
func (conn *MyWebsocketConn) Send(ctx context.Context, req writeReq) bool {
	select {
	case <-conn.connCtx.Done():
		return false
	case <-ctx.Done():
		return false
	case conn.writeCh <- req:
		return true
	}
}

// Close 关闭当前连接
func (conn *MyWebsocketConn) Close() {
	conn.closeOnce.Do(func() {
		conn.connCancel()
		_ = conn.conn.Close()
	})
}
