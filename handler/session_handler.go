package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

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

func (hdl *SessionHandler) GetSession(ctx *gin.Context) {
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

	// 获取会话
	sessionDTO, err := hdl.sessionSvc.GetSession(ctx, uid, targetID)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 返回
	response.Success(ctx, "获取会话成功", sessionDTO)
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

	hdl.sessionSvc.Message(ctx, uid, targetID)
}

// 获取历史消息
func (hdl *SessionHandler) GetHistoryMessage(ctx *gin.Context) {
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

	total, messageDTOs, err := hdl.sessionSvc.GetHistoryMessagesByPage(ctx, uid, targetID, pageNo, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	hasMore := (pageNo-1)*pageSize < total

	response.Success(ctx, "获取聊天记录成功", gin.H{
		"total":    total,
		"has_more": hasMore,
		"messages": messageDTOs,
	})
}
