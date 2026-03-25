package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	"github.com/yzletter/go-postery/microservice-backend/bff/conf"
	"github.com/yzletter/go-postery/microservice-backend/bff/dto/agent"
	"github.com/yzletter/go-postery/microservice-backend/bff/errno"
	grpcclient "github.com/yzletter/go-postery/microservice-backend/bff/grpc/client"
	"github.com/yzletter/go-postery/microservice-backend/bff/utils"
	"github.com/yzletter/go-postery/microservice-backend/bff/utils/response"
	"google.golang.org/grpc/codes"
)

type AgentHandler struct {
	agentSvc grpcclient.AgentClient
}

func NewAgentHandler(agentSvc grpcclient.AgentClient) *AgentHandler {
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
	var req agent.ChatAgentRequest
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
	resp, err := hdl.agentSvc.Chat(ctx, &agent_grpc.ChatRequest{UserID: uid, SessionID: ssid, Query: req.Query})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.Internal: errno.ErrServerInternal,
		}, errno.ErrServerInternal), agent.DTO{Documents: []string{}})
		return
	}

	// 返回响应
	response.Success(ctx, "success", agent.ToDTO(resp))
	return
}
