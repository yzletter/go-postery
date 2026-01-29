package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/bff/conf"
	userdto "github.com/yzletter/go-postery/bff/dto/user"
	"github.com/yzletter/go-postery/bff/utils"
	"github.com/yzletter/go-postery/bff/utils/response"
	"github.com/yzletter/go-postery/errno"
)

type UserHandler struct {
	userSvc user_grpc.UserServiceClient
}

// NewUserHandler 构造函数
func NewUserHandler(userSvc user_grpc.UserServiceClient) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
	}
}

// Profile 获取个人资料
func (hdl *UserHandler) Profile(ctx *gin.Context) {
	uid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	profile, err := hdl.userSvc.GetProfileById(ctx, &user_grpc.GetProfileByIdRequest{ID: uid})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取个人资料成功", userdto.ToDetailDTO(profile))
}

// ModifyProfile 修改个人资料
func (hdl *UserHandler) ModifyProfile(ctx *gin.Context) {
	var modifyProfileReq userdto.ModifyProfileRequest
	// 将请求参数绑定到结构体
	if err := ctx.ShouldBindJSON(&modifyProfileReq); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	_, err = hdl.userSvc.UpdateProfile(ctx, &user_grpc.UpdateProfileRequest{
		ID:       uid,
		Nickname: modifyProfileReq.Nickname,
		Avatar:   modifyProfileReq.Avatar,
		Bio:      modifyProfileReq.Bio,
		Gender:   uint32(modifyProfileReq.Gender),
		Birthday: modifyProfileReq.BirthDay,
		Location: modifyProfileReq.Location,
		Country:  modifyProfileReq.Country,
	})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, "修改个人资料成功", nil)
}

// Top 获取热门推荐用户
func (hdl *UserHandler) Top(ctx *gin.Context) {
	resp, err := hdl.userSvc.Top(ctx, &user_grpc.TopRequest{})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	res := make([]userdto.TopDTO, 0, len(resp.TopUsers))
	for _, u := range resp.TopUsers {
		top := userdto.ToTopDTO(u)
		res = append(res, top)
	}

	response.Success(ctx, "获取推荐关注成功", res)
}

func (hdl *UserHandler) Follow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	if uid == id {
		response.Error(ctx, errno.ErrFollowYourself)
		return
	}

	// 关注
	if _, err = hdl.userSvc.Follow(ctx, &user_grpc.FollowCommonRequest{FollowerID: uid, FolloweeID: id}); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "关注成功", nil)
}

func (hdl *UserHandler) UnFollow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	if uid == id {
		response.Error(ctx, errno.ErrFollowYourself)
		return
	}

	// 取消关注
	if _, err = hdl.userSvc.UnFollow(ctx, &user_grpc.FollowCommonRequest{FollowerID: uid, FolloweeID: id}); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "取消关注成功", nil)
}

func (hdl *UserHandler) IfFollow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	if uid == id {
		response.Error(ctx, errno.ErrFollowYourself)
		return
	}

	resp, err := hdl.userSvc.IfFollow(ctx, &user_grpc.FollowCommonRequest{FollowerID: uid, FolloweeID: id})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取关注关系成功", resp.Result)
}

// ListFollowers 返回关注我的人
func (hdl *UserHandler) ListFollowers(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	resp, err := hdl.userSvc.ListFollowersByPage(ctx, &user_grpc.ListFollowRequest{UserID: uid, PageNo: uint32(pageNo), PageSize: uint32(pageSize)})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	hasMore := pageNo*pageSize < int(resp.Count)

	response.Success(ctx, "获取粉丝列表成功",
		gin.H{
			"followers": userdto.BriefsToDTO(resp.UserBriefs),
			"total":     resp.Count,
			"hasMore":   hasMore,
		})
}

// ListFollowees 返回我关注的人
func (hdl *UserHandler) ListFollowees(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	resp, err := hdl.userSvc.ListFolloweesByPage(ctx, &user_grpc.ListFollowRequest{UserID: uid, PageNo: uint32(pageNo), PageSize: uint32(pageSize)})
	if err != nil {
		response.Error(ctx, err)
		return
	}

	hasMore := pageNo*pageSize < int(resp.Count)

	response.Success(ctx, "获取关注列表成功",
		gin.H{
			"followers": userdto.BriefsToDTO(resp.UserBriefs),
			"total":     resp.Count,
			"hasMore":   hasMore,
		})
}
