package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	agent2 "github.com/yzletter/go-postery/backend/bff/dto/agent"
	"github.com/yzletter/go-postery/backend/bff/errno"
	"github.com/yzletter/go-postery/backend/conf"
	grpcclient "github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/utils"
	"github.com/yzletter/go-postery/backend/utils/response"
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

func (hdl *AgentHandler) RegisterRouter(engine *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// 智能体模块
	agent := engine.Group("/agent")
	agent.Use(authMiddleware)
	{
		agent.POST("/chat", hdl.Chat) // POST /api/v1/agent/chat
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
	var req agent2.ChatAgentRequest
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
		}, errno.ErrServerInternal), agent2.DTO{Documents: []string{}})
		return
	}

	// 返回响应
	response.Success(ctx, "success", agent2.ToDTO(resp))
	return
}
