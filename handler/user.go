package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/infra/security"
	"github.com/yzletter/go-postery/repository/dao"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserService service.UserService
}

// NewUserHandler 构造函数
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

// Logout 用户登出 Handler
func (hdl *UserHandler) Logout(ctx *gin.Context) {
	// 设置 Cookie 里的双 Token 都置为 -1
	ctx.SetCookie(RefreshTokenInCookie, "", -1, "/", "localhost", false, true)
	ctx.SetCookie(AccessTokenInCookie, "", -1, "/", "localhost", false, true)
	response.Success(ctx, nil)
}

// ModifyPass 修改密码 Handler
func (hdl *UserHandler) ModifyPass(ctx *gin.Context) {
	var modifyPassRequest user.ModifyPassRequest
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&modifyPassRequest)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := service.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	err = hdl.UserService.UpdatePassword(ctx, uid, modifyPassRequest.OldPass, modifyPassRequest.NewPass)
	if err != nil {
		// 密码更改失败
		response.ServerError(ctx, "")
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, nil)
}

func (hdl *UserHandler) Profile(ctx *gin.Context) {
	uid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	ok, userDetailDTO := hdl.UserService.GetDetailById(ctx, int64(uid))
	if !ok {
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, userDetailDTO)
}

func (hdl *UserHandler) ModifyProfile(ctx *gin.Context) {
	var modifyUserProfileRequest user.ModifyProfileRequest
	// 将请求参数绑定到结构体
	err := ctx.ShouldBind(&modifyUserProfileRequest)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := service.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	err = hdl.UserService.UpdateProfile(ctx, uid, modifyUserProfileRequest)
	if err != nil {
		slog.Error("Error", err)
		if errors.Is(err, dao.ErrRecordNotFound) {
			response.ParamError(ctx, "")
			return
		}

		response.ServerError(ctx, "")
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, nil)
}
