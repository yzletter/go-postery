package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/conf"
	grpcclient "github.com/yzletter/go-postery/backend/grpc/manager"
	sessiondto "github.com/yzletter/go-postery/backend/micro/bff/dto/session"
	"github.com/yzletter/go-postery/backend/micro/bff/errno"
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
	for _, session := range resp.Sessions {
		user, err := hdl.userSvc.GetProfileById(ctx, &user_grpc.GetProfileByIdRequest{ID: session.TargetID})
		if err != nil {
			user = &user_grpc.UserDetail{}
		}
		sessions = append(sessions, sessiondto.ToSessionDTO(session, user))
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
	session, err := hdl.sessionSvc.GetSession(ctx, &session_grpc.BothUserID{UserID: uid, TargetID: targetID})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), sessiondto.SessionDTO{})
		return
	}

	// 获取用户
	user, err := hdl.userSvc.GetProfileById(ctx, &user_grpc.GetProfileByIdRequest{ID: session.TargetID})
	if err != nil {
		user = &user_grpc.UserDetail{}
	}

	// 返回
	response.Success(ctx, "获取会话成功", sessiondto.ToSessionDTO(session, user))
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

	hasMore := (pageNo-1)*pageSize < int(resp.Count)

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
