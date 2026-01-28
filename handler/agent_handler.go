package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/auth/conf"
	agentdto "github.com/yzletter/go-postery/dto/agent"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type AgentHandler struct {
	agentSvc service.AgentService
}

func NewAgentHandler(agentSvc service.AgentService) *AgentHandler {
	return &AgentHandler{
		agentSvc: agentSvc,
	}
}

func (hdl *AgentHandler) Chat(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取参数并校验
	var req agentdto.ChatAgentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	ssid, err := strconv.ParseInt(req.SessionID, 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 询问 Agent
	resp, err := hdl.agentSvc.Chat(ctx, uid, ssid, req.Query)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 返回响应
	response.Success(ctx, "success", resp)
	return
}
