package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type FollowHandler struct {
	followSvc service.FollowService
	userSvc   service.UserService
}

func NewFollowHandler(followSvc service.FollowService, userSvc service.UserService) *FollowHandler {
	return &FollowHandler{
		followSvc: followSvc,
		userSvc:   userSvc,
	}
}

func (hdl *FollowHandler) Follow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取对方 id
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 关注
	err = hdl.followSvc.Follow(ctx, uid, id)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "关注成功", nil)
}

func (hdl *FollowHandler) UnFollow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取对方 id
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 取消关注
	err = hdl.followSvc.UnFollow(ctx, uid, id)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "取消关注成功", nil)
}

func (hdl *FollowHandler) IfFollow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取对方 id
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	res, err := hdl.followSvc.IfFollow(ctx, uid, id)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取关注关系成功", res)
}

// ListFollowers 返回关注我的人
func (hdl *FollowHandler) ListFollowers(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil || pageNo < 1 || pageSize > 100 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	total, followerDTOs, err := hdl.followSvc.ListFollowersByPage(ctx, uid, pageNo, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	hasMore := pageNo*pageSize < total

	response.Success(ctx, "获取粉丝列表成功", gin.H{
		"followers": followerDTOs,
		"total":     total,
		"hasMore":   hasMore,
	})
}

// ListFollowees 返回我关注的人
func (hdl *FollowHandler) ListFollowees(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil || pageNo < 1 || pageSize > 100 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	total, followerDTOs, err := hdl.followSvc.ListFolloweesByPage(ctx, uid, pageNo, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	hasMore := pageNo*pageSize < total

	response.Success(ctx, "获取关注列表成功", gin.H{
		"followees": followerDTOs,
		"total":     total,
		"hasMore":   hasMore,
	})
}
