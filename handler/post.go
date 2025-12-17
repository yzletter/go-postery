package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/dto/post"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type PostHandler struct {
	PostService service.PostService
	UserService service.UserService
	TagSvc      service.TagService
}

func NewPostHandler(postService service.PostService, userService service.UserService, tagSvc service.TagService) *PostHandler {
	return &PostHandler{
		PostService: postService,
		UserService: userService,
		TagSvc:      tagSvc,
	}
}

// List 获取帖子列表
func (hdl *PostHandler) List(ctx *gin.Context) {
	// 从 /posts?pageNo=1&pageSize=2 路由中拿出 pageNo 和 pageSize
	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil {
		// 获取帖子列表请求的参数不合法
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 获取帖子总数和当前页帖子列表
	total, postDTOs := hdl.PostService.ListByPage(ctx, pageNo, pageSize)
	for k := range postDTOs {
		postDTOs[k].Tags = hdl.TagSvc.FindTagsByPostID(int(postDTOs[k].ID))
	}

	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := pageNo*pageSize < total

	// 返回
	response.Success(ctx, gin.H{
		"posts":   postDTOs,
		"total":   total,
		"hasMore": hasMore,
	})
	return
}

// ListByTag 根据标签获取帖子列表
func (hdl *PostHandler) ListByTag(ctx *gin.Context) {
	// 从 /posts?pageNo=1&pageSize=2&tag= 路由中拿出 pageNo 和 pageSize
	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	name := ctx.Query("tag")
	if err1 != nil || err2 != nil {
		// 获取帖子列表请求的参数不合法
		response.ParamError(ctx, "")
		return
	}

	// 获取帖子总数和当前页帖子列表
	total, postDTOs := hdl.PostService.GetByPageAndTag(ctx, name, pageNo, pageSize)
	for k := range postDTOs {
		postDTOs[k].Tags = hdl.TagSvc.FindTagsByPostID(int(postDTOs[k].ID))
	}

	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := hdl.PostService.HasMore(ctx, pageNo, pageSize, total)

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
	ok, postDTO := hdl.PostService.GetDetailById(ctx, pid)
	if !ok {
		response.ServerError(ctx, "")
		return
	}
	postDTO.Tags = hdl.TagSvc.FindTagsByPostID(int(postDTO.ID))
	response.Success(ctx, postDTO)
}

// Create 创建帖子
func (hdl *PostHandler) Create(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 参数绑定
	var createRequest post.CreateRequest
	err = ctx.ShouldBindJSON(&createRequest)
	if err != nil {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	// 创建帖子
	postDTO, err := hdl.PostService.Create(ctx, int(uid), createRequest.Title, createRequest.Content)
	if err != nil {
		// 创建帖子失败
		response.ServerError(ctx, "")
		return
	}

	// 建立标签
	hdl.TagSvc.Bind(int(postDTO.ID), createRequest.Tags)

	response.Success(ctx, postDTO)
}

// Delete 删除帖子
func (hdl *PostHandler) Delete(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx)
	if err != nil {
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
	err = hdl.PostService.Delete(ctx, pid, int(uid))
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
	uid, err := utils.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 参数绑定
	var updateRequest post.UpdateRequest
	err = ctx.ShouldBindJSON(&updateRequest)

	if err != nil || updateRequest.Id == 0 {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	// 修改
	err = hdl.PostService.Update(ctx, updateRequest.Id, int(uid), updateRequest.Title, updateRequest.Content, updateRequest.Tags)
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
	uid, err := utils.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 判断登录用户是否是作者
	ok := hdl.PostService.Belong(ctx, pid, int(uid))
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

	postDTOs := hdl.PostService.GetByUid(ctx, uid)
	response.Success(ctx, postDTOs)
}

func (hdl *PostHandler) Like(ctx *gin.Context) {
	// 从路由中获取帖子 id
	pid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	// 从 CTX 中获取 uid
	uid, err := utils.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	err = hdl.PostService.Like(ctx, pid, int(uid))
	if err != nil {
		if errors.Is(err, userLikeRepository.ErrRecordHasExist) {
			response.Fail(ctx, response.CodeBadRequest, "重复点赞")
			return
		}
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, "")
}

func (hdl *PostHandler) Dislike(ctx *gin.Context) {
	// 从路由中获取帖子 id
	pid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	// 从 CTX 中获取 uid
	uid, err := utils.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	err = hdl.PostService.Dislike(ctx, pid, int(uid))
	if err != nil {
		if errors.Is(err, userLikeRepository.ErrRecordNotExist) {
			response.Fail(ctx, response.CodeBadRequest, "重复取消")
			return
		}
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, "")
}

func (hdl *PostHandler) IfLike(ctx *gin.Context) {
	// 从路由中获取帖子 id
	pid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	// 从 CTX 中获取 uid
	uid, err := utils.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	ok, err := hdl.PostService.IfLike(pid, int(uid))

	response.Success(ctx, ok)
}
