package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	sessiondto "github.com/yzletter/go-postery/backend/bff/dto/session"
	"github.com/yzletter/go-postery/backend/bff/errno"
	"github.com/yzletter/go-postery/backend/conf"
	grpcclient "github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/utils"
	"github.com/yzletter/go-postery/backend/utils/response"
	"google.golang.org/grpc/codes"
)

type SessionHandler struct {
	sessionSvc grpcclient.SessionClient
	userSvc    grpcclient.UserClient
}

func NewSessionHandler(sessionSvc grpcclient.SessionClient, userSvc grpcclient.UserClient) *SessionHandler {
	return &SessionHandler{
		sessionSvc: sessionSvc,
		userSvc:    userSvc,
	}
}

func (hdl *SessionHandler) RegisterRouter(engine *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// 私信模块
	sessions := engine.Group("/sessions")
	sessions.Use(authMiddleware)
	{
		sessions.GET("", hdl.List)               // GET /api/v1/sessions													获取当前登录用户会话列表
		sessions.POST("/:id/delete", hdl.Delete) // POST /api/v1/sessions/:id/delete										删除当前会话
	}

	chat := sessions.Group("/target/:id")
	{
		chat.GET("", hdl.GetSession)                 // GET /api/v1/sessions/target/:id/									获取与对方的会话
		chat.GET("/messages", hdl.GetHistoryMessage) // GET /api/v1/sessions/target/:id/messages?pageNo=1&pageSize=5		获取与对方会话的聊天记录
	}

}

// List 列出会话列表
func (hdl *SessionHandler) List(ctx *gin.Context) {
	// 取当前登录用户 uid
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 列出当前用户的会话列表
	resp, err := hdl.sessionSvc.ListByUID(ctx, &session_grpc.UserID{UserID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), []sessiondto.SessionDTO{})
		return
	}

	sessions := make([]sessiondto.SessionDTO, 0, len(resp.Sessions))
	for _, sess := range resp.Sessions {
		user, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: sess.TargetID})
		if err != nil {
			user = &user_grpc.Profile{}
		}
		sessions = append(sessions, sessiondto.ToSessionDTO(sess, user))
	}

	// 返回
	response.Success(ctx, "获取会话列表成功", sessions)
}

func (hdl *SessionHandler) GetSession(ctx *gin.Context) {
	// 取当前登录用户 uid
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	// 获取会话
	sess, err := hdl.sessionSvc.GetSession(ctx, &session_grpc.BothUserID{UserID: uid, TargetID: targetID})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), sessiondto.SessionDTO{})
		return
	}

	// 获取用户
	user, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: sess.TargetID})
	if err != nil {
		user = &user_grpc.Profile{}
	}

	// 返回
	response.Success(ctx, "获取会话成功", sessiondto.ToSessionDTO(sess, user))
}

func (hdl *SessionHandler) Delete(ctx *gin.Context) {
	// 取当前登录用户 uid
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 从 URL 中获取 SessionID
	sid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 删除会话

	if _, err = hdl.sessionSvc.Delete(ctx, &session_grpc.DeleteRequest{UserID: uid, SessionID: sid}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.Unauthenticated: errno.ErrUnauthorized,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "删除会话成功", nil)
	return
}

// GetHistoryMessage 获取历史消息
func (hdl *SessionHandler) GetHistoryMessage(ctx *gin.Context) {
	// 取当前登录用户 uid
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	// 取 pageNo 和 pageSize
	pageNo, err := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}
	pageSize, err := strconv.Atoi(ctx.DefaultQuery("pageSize", "5"))
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	resp, err := hdl.sessionSvc.GetHistoryMessagesByPage(ctx, &session_grpc.GetHistoryMessagesByPageRequest{
		UserID: uid, TargetID: targetID, PageNo: uint32(pageNo), PageSize: uint32(pageSize)})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), gin.H{
			"messages": []sessiondto.MessageDTO{},
			"total":    0,
			"has_more": false,
		})
		return
	}

	hasMore := pageNo*pageSize < int(resp.Count)

	messages := make([]sessiondto.MessageDTO, 0, len(resp.Messages))
	for _, message := range resp.Messages {
		messages = append(messages, sessiondto.ToMessageDTO(message))
	}

	response.Success(ctx, "获取聊天记录成功", gin.H{
		"messages": messages,
		"total":    resp.Count,
		"has_more": hasMore,
	})
}
