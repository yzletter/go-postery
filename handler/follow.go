package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils/response"
)

type FollowHandler struct {
	FollowSvc *service.FollowService
	UserSvc   *service.UserService
}

func NewFollowHandler(followSvc *service.FollowService, userSvc *service.UserService) *FollowHandler {
	return &FollowHandler{FollowSvc: followSvc, UserSvc: userSvc}
}

func (hdl *FollowHandler) Follow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := service.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 对方 id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	err = hdl.FollowSvc.Follow(uid, id)
	if err != nil {
		if errors.Is(err, service.ErrDuplicatedFollow) {
			response.Fail(ctx, response.CodeBadRequest, "重复关注")
			return
		}
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, "")
}

func (hdl *FollowHandler) DisFollow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := service.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 对方 id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	err = hdl.FollowSvc.DisFollow(uid, id)
	if err != nil {
		if errors.Is(err, service.ErrDuplicatedDisFollow) {
			response.Fail(ctx, response.CodeBadRequest, "重复取消关注")
			return
		}
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, "")
}

func (hdl *FollowHandler) IfFollow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := service.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 对方 id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	res, err := hdl.FollowSvc.IfFollow(uid, id)
	if err != nil {
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, res)
}

// ListFollowers 返回关注我的人
func (hdl *FollowHandler) ListFollowers(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := service.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	followerDTOs, err := hdl.FollowSvc.GetFollowers(uid)
	if err != nil {
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, followerDTOs)
}

// ListFollowees 返回被我关注的人
func (hdl *FollowHandler) ListFollowees(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := service.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	followerDTOs, err := hdl.FollowSvc.GetFollowees(uid)
	if err != nil {
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, followerDTOs)
}
