package handler

import (
	"fmt"
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
	total, postDTOs, err := hdl.PostService.ListByPage(ctx, pageNo, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	for k := range postDTOs {
		res, err := hdl.TagSvc.FindTagsByPostID(ctx, postDTOs[k].ID)
		if err != nil {
			continue
		}
		postDTOs[k].Tags = res
	}

	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := pageNo*pageSize < total

	// 返回
	response.Success(ctx, "获取帖子列表成功", gin.H{
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
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 获取帖子总数和当前页帖子列表
	total, postDTOs, err := hdl.PostService.ListByPageAndTag(ctx, name, pageNo, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	for k := range postDTOs {
		res, err := hdl.TagSvc.FindTagsByPostID(ctx, postDTOs[k].ID)
		if err != nil {
			continue
		}
		postDTOs[k].Tags = res
	}

	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := pageNo*pageSize < total

	// 返回
	response.Success(ctx, "获取帖子列表成功", gin.H{
		"posts":   postDTOs,
		"total":   total,
		"hasMore": hasMore,
	})
	return
}

// Detail 获取帖子详情
func (hdl *PostHandler) Detail(ctx *gin.Context) {
	// 从路由中获取 pid 参数
	pid, err := strconv.ParseInt(ctx.Param("pid"), 10, 64)
	if err != nil {
		// 获取帖子详情请求的参数不合法
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 根据 pid 查找帖子详情
	postDTO, err := hdl.PostService.GetDetailById(ctx, pid, true)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	postDTO.Tags, err = hdl.TagSvc.FindTagsByPostID(ctx, postDTO.ID)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取帖子详情成功", postDTO)
}

// Create 创建帖子
func (hdl *PostHandler) Create(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 参数绑定
	var createRequest post.CreateRequest
	err = ctx.ShouldBindJSON(&createRequest)
	if err != nil {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 创建帖子
	postDTO, err := hdl.PostService.Create(ctx, uid, createRequest.Title, createRequest.Content)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 建立标签
	err = hdl.TagSvc.Bind(ctx, postDTO.ID, createRequest.Tags)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "帖子创建成功", postDTO)
}

// Delete 删除帖子
func (hdl *PostHandler) Delete(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 再拿帖子 pid
	pid, err := strconv.ParseInt(ctx.Param("pid"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 进行删除
	err = hdl.PostService.Delete(ctx, pid, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "帖子删除成功", nil)
}

// Update 修改帖子
func (hdl *PostHandler) Update(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 参数绑定
	var updateRequest post.UpdateRequest
	err = ctx.ShouldBindJSON(&updateRequest)

	if err != nil || updateRequest.ID == 0 {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 修改
	err = hdl.PostService.Update(ctx, updateRequest.ID, uid, updateRequest.Title, updateRequest.Content, updateRequest.Tags)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "帖子更新成功", nil)
	return
}

// Belong 查询帖子作者是否为当前登录用户
func (hdl *PostHandler) Belong(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取帖子 id
	pid, err := strconv.ParseInt(ctx.Param("pid"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 判断登录用户是否是作者
	ok := hdl.PostService.Belong(ctx, pid, uid)
	if !ok {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	response.Success(ctx, "", nil)
	return
}

// ListByPageAndUid 按页获取目标用户发布的帖子
func (hdl *PostHandler) ListByPageAndUid(ctx *gin.Context) {
	// 从路由中获取 uid
	uid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		fmt.Println(uid)
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil || pageNo < 1 || pageSize > 100 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	total, postDTOs, err := hdl.PostService.ListByPageAndUid(ctx, uid, pageNo, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := pageNo*pageSize < total

	// 返回
	response.Success(ctx, "获取帖子列表成功", gin.H{
		"posts":   postDTOs,
		"total":   total,
		"hasMore": hasMore,
	})
}

func (hdl *PostHandler) Like(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取帖子 id
	pid, err := strconv.ParseInt(ctx.Param("pid"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	err = hdl.PostService.Like(ctx, pid, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "", nil)
}

func (hdl *PostHandler) Unlike(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取帖子 id
	pid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	err = hdl.PostService.Unlike(ctx, pid, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "", nil)
}

func (hdl *PostHandler) IfLike(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取帖子 id
	pid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	ok, err := hdl.PostService.IfLike(ctx, pid, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "", ok)
}
