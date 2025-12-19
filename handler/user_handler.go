package handler

import (
	"log/slog"
	"strconv"

	"github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc service.UserService
}

// NewUserHandler 构造函数
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userSvc: userService,
	}
}

// ModifyPass 修改密码 Handler
func (hdl *UserHandler) ModifyPass(ctx *gin.Context) {
	var modifyPassReq user.ModifyPassRequest
	// 将请求参数绑定到结构体
	err := ctx.ShouldBindJSON(&modifyPassReq)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	err = hdl.userSvc.UpdatePassword(ctx, uid, modifyPassReq.OldPass, modifyPassReq.NewPass)
	if err != nil {
		// 密码更改失败
		response.Error(ctx, err)
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, "密码修改成功", nil)
}

func (hdl *UserHandler) Profile(ctx *gin.Context) {
	uid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	userDetailDTO, err := hdl.userSvc.GetDetailById(ctx, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取个人资料成功", userDetailDTO)
}

func (hdl *UserHandler) ModifyProfile(ctx *gin.Context) {
	var modifyProfileReq user.ModifyProfileRequest
	// 将请求参数绑定到结构体
	err := ctx.ShouldBindJSON(&modifyProfileReq)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	err = hdl.userSvc.UpdateProfile(ctx, uid, modifyProfileReq)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, "修改个人资料成功", nil)
}
