package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
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

type SessionHandler struct {
	sessionSvc service.SessionService
}

func NewSessionHandler(sessionSvc service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionSvc: sessionSvc,
	}
}

// List 列出会话列表
func (hdl *SessionHandler) List(ctx *gin.Context) {
	// 取当前登录用户 uid
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 列出当前用户的会话列表
	sessionDTOs, err := hdl.sessionSvc.ListByUid(ctx, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 返回
	response.Success(ctx, "获取会话列表成功", sessionDTOs)
}

func (hdl *SessionHandler) Delete(ctx *gin.Context) {
	// 取当前登录用户 uid
	_, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 从 URL 中获取 SessionID
	_, err = strconv.ParseInt(ctx.Param("sid"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	response.Success(ctx, "删除会话成功", nil)
	return
}

func (hdl *SessionHandler) MessageToUser(ctx *gin.Context) {
	// 取当前登录用户 uid
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 取对方 target_id
	targetID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 升级 HTTP
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 心跳保持
	go heartBeat(conn)

	err = hdl.sessionSvc.Message(ctx, conn, uid, targetID)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "", nil)
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
