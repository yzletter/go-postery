package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type WebsocketHandler struct {
	websocketSvc service.WebsocketService
}

func NewWebsocketHandler(websocketSvc service.WebsocketService) *WebsocketHandler {
	return &WebsocketHandler{websocketSvc: websocketSvc}
}

func (hdl *WebsocketHandler) Connect(ctx *gin.Context) {
	// 取当前登录用户 uid
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	if err := hdl.websocketSvc.Connect(ctx, ctx.Writer, ctx.Request, uid); err != nil {
		response.Error(ctx, err)
		return
	}
}
