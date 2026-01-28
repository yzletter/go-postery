package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/auth/conf"
	"github.com/yzletter/go-postery/bff/dto/user"
	utils2 "github.com/yzletter/go-postery/bff/utils"
	"github.com/yzletter/go-postery/bff/utils/response"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
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

// Profile 获取个人资料
func (hdl *UserHandler) Profile(ctx *gin.Context) {
	uid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	userDetailDTO, err := hdl.userSvc.GetProfileById(ctx, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取个人资料成功", userDetailDTO)
}

// ModifyProfile 修改个人资料
func (hdl *UserHandler) ModifyProfile(ctx *gin.Context) {
	var modifyProfileReq user.ModifyProfileRequest
	// 将请求参数绑定到结构体
	err := ctx.ShouldBindJSON(&modifyProfileReq)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils2.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
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

// Top 获取热门推荐用户
func (hdl *UserHandler) Top(ctx *gin.Context) {
	userDTOs, err := hdl.userSvc.Top(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取推荐关注成功", userDTOs)
}
