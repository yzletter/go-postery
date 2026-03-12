package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/microservice-backend/bff/conf"
	"github.com/yzletter/go-postery/microservice-backend/bff/errno"
	"github.com/yzletter/go-postery/microservice-backend/bff/service"
	"github.com/yzletter/go-postery/microservice-backend/bff/utils"
	"github.com/yzletter/go-postery/microservice-backend/bff/utils/response"
)

type WebsocketHandler struct {
	websocketSvc service.WebsocketService
}

func NewWebsocketHandler(websocketSvc service.WebsocketService) *WebsocketHandler {
	return &WebsocketHandler{websocketSvc: websocketSvc}
}

func (hdl *WebsocketHandler) Connect(ctx *gin.Context) {
	// 取当前登录用户 uid
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	if err := hdl.websocketSvc.Connect(ctx, ctx.Writer, ctx.Request, uid); err != nil {
		response.Error(ctx, err)
		return
	}
}
