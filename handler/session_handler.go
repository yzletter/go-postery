package handler

import (
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
