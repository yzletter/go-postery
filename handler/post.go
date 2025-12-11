package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/dto/request"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type PostHandler struct {
	PostService *service.PostService
	UserService *service.UserService
}

func NewPostHandler(postService *service.PostService, userService *service.UserService) *PostHandler {
	return &PostHandler{
		PostService: postService,
		UserService: userService,
	}
}

// List 获取帖子列表
func (hdl *PostHandler) List(ctx *gin.Context) {
	// 从 /posts?pageNo=1&pageSize=2 路由中拿出 pageNo 和 pageSize
	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil {
		// 获取帖子列表请求的参数不合法
		response.ParamError(ctx, "")
		return
	}

	// 获取帖子总数和当前页帖子列表
	total, postDTOs := hdl.PostService.GetByPage(pageNo, pageSize)
	slog.Info("", "postDTOs", postDTOs)

	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := hdl.PostService.HasMore(pageNo, pageSize, total)

	// 返回
	response.Success(ctx, gin.H{
		"posts":   postDTOs,
		"total":   total,
		"hasMore": hasMore,
	})

	return
}

// Detail 获取帖子详情
func (hdl *PostHandler) Detail(ctx *gin.Context) {
	// 从路由中获取 pid 参数
	pid, err := strconv.Atoi(ctx.Param("pid"))
	if err != nil {
		// 获取帖子详情请求的参数不合法
		response.ParamError(ctx, "")
		return
	}

	// 根据 pid 查找帖子详情
	ok, postDTO := hdl.PostService.GetDetailById(pid)
	if !ok {
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, postDTO)
}

// Create 创建帖子
func (hdl *PostHandler) Create(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := strconv.ParseInt(ctx.Value(service.UID_IN_CTX).(string), 10, 64)
	if err != nil {
		// 没有登录
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 参数绑定
	var createRequest request.CreatePostRequest
	err = ctx.ShouldBind(&createRequest)
	if err != nil {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	// 创建帖子
	postDTO, err := hdl.PostService.Create(int(uid), createRequest.Title, createRequest.Content)
	if err != nil {
		// 创建帖子失败
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, postDTO)
}

// Delete 删除帖子
func (hdl *PostHandler) Delete(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := strconv.ParseInt(ctx.Value(service.UID_IN_CTX).(string), 10, 64)
	if err != nil {
		// 没有登录
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 再拿帖子 pid
	pid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || pid == 0 {
		response.ParamError(ctx, "")
		return
	}

	// 进行删除
	err = hdl.PostService.Delete(pid, int(uid))
	if err != nil {
		if err.Error() == "没有权限" {
			response.Unauthorized(ctx, "")
		}
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, nil)
}

// Update 修改帖子
func (hdl *PostHandler) Update(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := strconv.ParseInt(ctx.Value(service.UID_IN_CTX).(string), 10, 64)
	if err != nil {
		// 没有登录
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 参数绑定
	var updateRequest request.UpdatePostRequest
	err = ctx.ShouldBind(&updateRequest)

	if err != nil || updateRequest.Id == 0 {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	// 修改
	err = hdl.PostService.Update(updateRequest.Id, int(uid), updateRequest.Title, updateRequest.Content)
	if err != nil {
		if err.Error() == "没有权限" {
			response.Unauthorized(ctx, "")
		}
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, nil)
	return
}

// Belong 查询帖子作者是否为当前登录用户
func (hdl *PostHandler) Belong(ctx *gin.Context) {
	// 获取帖子 id
	pid, err := strconv.Atoi(ctx.Query("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := strconv.ParseInt(ctx.Value(service.UID_IN_CTX).(string), 10, 64)
	if err != nil {
		// 没有登录
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 判断登录用户是否是作者
	ok := hdl.PostService.Belong(pid, int(uid))
	if !ok {
		response.Unauthorized(ctx, "")
		return
	}

	// 属于
	response.Success(ctx, "")
	return
}

// ListByUid 获取目标用户发布的帖子
func (hdl *PostHandler) ListByUid(ctx *gin.Context) {
	// 从路由中获取 uid
	uid, err := strconv.Atoi(ctx.Param("uid"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	postDTOs := hdl.PostService.GetByUid(uid)
	response.Success(ctx, postDTOs)
}
