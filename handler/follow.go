package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/service"
)

type FollowHandler struct {
	FollowSvc *service.FollowService
	UserSvc   *service.UserService
}

func NewFollowHandler(followSvc *service.FollowService, userSvc *service.UserService) *FollowHandler {
	return &FollowHandler{FollowSvc: followSvc, UserSvc: userSvc}
}

func (hdl *FollowHandler) Follow(ctx *gin.Context) {

}

func (hdl *FollowHandler) DisFollow(ctx *gin.Context) {

}

func (hdl *FollowHandler) IfFollow(ctx *gin.Context) {

}

func (hdl *FollowHandler) ListFollowers(ctx *gin.Context) {

}

func (hdl *FollowHandler) ListFollowees(ctx *gin.Context) {

}
