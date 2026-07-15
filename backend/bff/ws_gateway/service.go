package ws_gateway

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yzletter/go-postery/backend/bff/errno"
	"github.com/yzletter/go-postery/backend/bff/ws_gateway/domain"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/utils"
	"github.com/yzletter/go-postery/backend/utils/response"
)

// WSMessageHandler 用来处理客户端发来的业务消息
type WSMessageHandler interface {
	// NewSessionConnection 处理新 Session 连接
	NewSessionConnection(ctx context.Context, userID int64) error
	// HandleWSMessage 处理客户端发来的业务消息。
	HandleWSMessage(ctx context.Context, userID int64, msg WSMessage) error
}

// WSMessage 表示客户端发给服务端的消息结构
type WSMessage struct {
	BizType string      `json:"biz_type"`
	BizData interface{} `json:"biz_data"`
}

const (
	WSBizTypeSession   = "session"
	WSBizTypeInterview = "interview"
)

// HTTP 升级器：负责把普通 HTTP 请求升级为 WebSocket 连接
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 20 * time.Second,
	ReadBufferSize:   100000, // 读缓冲区大小
	WriteBufferSize:  100000, // 写缓冲区大小

	// CheckOrigin 用于校验跨域来源，防止任意站点建立 WebSocket 连接
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		if origin == "" {
			// 没有 Origin 的请求，一般来自 Postman、curl、命令行测试工具
			// 这里选择放行
			return true
		}

		// 前端地址白名单
		// 生产环境建议放到配置文件中，不要硬编码
		allowList := map[string]bool{
			"http://localhost:5173": true,
		}

		if allowList[origin] {
			return true
		}

		// 如果是同源请求，也放行
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}

		return strings.EqualFold(u.Host, r.Host)
	},
}

// WebsocketGateway WebSocket 网关
type WebsocketGateway struct {
	pool    sync.Map
	handler WSMessageHandler
}

// NewWebsocketGateway 创建 WebSocket 网关
func NewWebsocketGateway(handler WSMessageHandler) *WebsocketGateway {
	return &WebsocketGateway{
		handler: handler,
	}
}

// NewSessionConnectionGin gin HandlerFunc
func (gate *WebsocketGateway) NewSessionConnectionGin(ctx *gin.Context) {
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	if err := gate.newSessionConnection(ctx, uid, ctx.Writer, ctx.Request); err != nil {
		return
	}
}

// NewInterviewConnectionGin gin HandlerFunc
func (gate *WebsocketGateway) NewInterviewConnectionGin(ctx *gin.Context) {
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	if err := gate.newInterviewConnection(ctx, uid, ctx.Writer, ctx.Request); err != nil {
		return
	}
}

// newSessionConnection 创建一个新的聊天连接
func (gate *WebsocketGateway) newSessionConnection(ctx context.Context, userID int64, w http.ResponseWriter, r *http.Request) error {
	// 将 HTTP 请求升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	connCtx, cancel := context.WithCancel(context.Background())

	wsConn := &MyWebsocketConn{
		biz:        domain.BizSession,
		userID:     userID,
		gate:       gate,
		conn:       conn,
		writeCh:    make(chan writeReq, 50),
		connCtx:    connCtx,
		connCancel: cancel,
		closeOnce:  sync.Once{},
	}

	if err := gate.SaveConnection(ctx, userID, domain.BizSession, wsConn); err != nil {
		wsConn.Close()
		return err
	}

	// 写
	go wsConn.Writer()

	// 读
	go wsConn.Reader()

	// 心跳
	go wsConn.HeartBeat()

	// 启动与 WebSocket 同生命周期的 Session 消息消费任务。
	if gate.handler != nil {
		go func() {
			err := gate.handler.NewSessionConnection(connCtx, userID)
			if connCtx.Err() != nil {
				return
			}
			if err != nil {
				slog.Error("Websocket connection task stopped", "userID", userID, "error", err)
			}
			// 消费任务提前退出后关闭连接，避免保留只能发不能收的半连接。
			wsConn.Close()
		}()
	}

	return nil
}

// newInterviewConnection 创建一个新的面试连接
func (gate *WebsocketGateway) newInterviewConnection(ctx context.Context, userID int64, w http.ResponseWriter, r *http.Request) error {
	// 将 HTTP 请求升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	connCtx, cancel := context.WithCancel(context.Background())

	wsConn := &MyWebsocketConn{
		biz:        domain.BizInterview,
		userID:     userID,
		gate:       gate,
		conn:       conn,
		writeCh:    make(chan writeReq, 50),
		connCtx:    connCtx,
		connCancel: cancel,
		closeOnce:  sync.Once{},
	}

	if err := gate.SaveConnection(ctx, userID, domain.BizInterview, wsConn); err != nil {
		wsConn.Close()
		return err
	}

	// 写
	go wsConn.Writer()

	// 读
	go wsConn.Reader()

	// 心跳
	go wsConn.HeartBeat()

	return nil
}

// Push 给指定用户的连接推送消息, 这里的 data 是 WSMessage
func (gate *WebsocketGateway) Push(ctx context.Context, userID int64, biz domain.ConnType, data []byte) error {
	conn, err := gate.LoadConnection(ctx, userID, biz)
	if err != nil {
		return err
	}

	if ok := conn.Send(ctx, writeReq{messageType: websocket.TextMessage, data: data}); !ok {
		return errors.New("websocket send failed")
	}
	return nil
}

// SaveConnection 保存
func (gate *WebsocketGateway) SaveConnection(_ context.Context, userID int64, biz domain.ConnType, conn *MyWebsocketConn) error {
	// 如果旧连接存在，先关闭旧连接
	if oldVal, ok := gate.pool.Load(newKey(userID, biz)); ok {
		if oldConn, ok := oldVal.(*MyWebsocketConn); ok {
			oldConn.Close()
		}
	}
	// 保存新连接
	gate.pool.Store(userID, conn)
	return nil
}

// LoadConnection 根据 userID 获取连接
func (gate *WebsocketGateway) LoadConnection(_ context.Context, userID int64, biz domain.ConnType) (*MyWebsocketConn, error) {
	val, ok := gate.pool.Load(newKey(userID, biz))
	if !ok {
		return nil, errors.New("user not found")
	}

	conn, ok := val.(*MyWebsocketConn)
	if !ok {
		return nil, errors.New("conn not found")
	}

	return conn, nil
}

// DelConnection 删除连接
func (gate *WebsocketGateway) DelConnection(_ context.Context, userID int64, biz domain.ConnType, target *MyWebsocketConn) error {
	val, ok := gate.pool.Load(newKey(userID, biz))
	if !ok {
		return nil
	}

	cur, ok := val.(*MyWebsocketConn)
	if !ok {
		// 类型异常，直接删除脏数据
		gate.pool.Delete(newKey(userID, biz))
		return nil
	} else if cur != target {
		// 如果当前连接不是要删除的连接，说明用户已经重连了别的连接
		return nil
	}

	// 删除 userID -> conn 映射
	gate.pool.Delete(newKey(userID, biz))

	// 关闭连接
	cur.Close()

	return nil
}

func newKey(uid int64, biz domain.ConnType) string {
	return strconv.FormatInt(uid, 10) + "_" + string(biz)
}
