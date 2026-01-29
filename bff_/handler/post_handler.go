package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/auth/conf"
	"github.com/yzletter/go-postery/bff/errno"
	"github.com/yzletter/go-postery/bff_/dto/comment"
	utils2 "github.com/yzletter/go-postery/bff_/utils"
	"github.com/yzletter/go-postery/bff_/utils/response"
	"github.com/yzletter/go-postery/post/dto"
	"github.com/yzletter/go-postery/post/service"
)

type PostHandler struct {
	postSvc service.PostService
	userSvc service.UserService
	tagSvc  service.TagService
}

func NewPostHandler(postService service.PostService, userService service.UserService, tagSvc service.TagService) *PostHandler {
	return &PostHandler{
		postSvc: postService,
		userSvc: userService,
		tagSvc:  tagSvc,
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
	total, postDTOs, err := hdl.postSvc.ListByPage(ctx, pageNo, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	for k := range postDTOs {
		res, err := hdl.tagSvc.FindTagsByPostID(ctx, postDTOs[k].ID)
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

// ListByTagAndPage 根据标签获取帖子列表
func (hdl *PostHandler) ListByTagAndPage(ctx *gin.Context) {
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
	total, postDTOs, err := hdl.postSvc.ListByPageAndTag(ctx, name, pageNo, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	for k := range postDTOs {
		res, err := hdl.tagSvc.FindTagsByPostID(ctx, postDTOs[k].ID)
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
	pid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		// 获取帖子详情请求的参数不合法
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 根据 pid 查找帖子详情
	postDTO, err := hdl.postSvc.GetDetailById(ctx, pid, true)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	postDTO.Tags, err = hdl.tagSvc.FindTagsByPostID(ctx, postDTO.ID)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取帖子详情成功", postDTO)
}

// Create 创建帖子
func (hdl *PostHandler) Create(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 参数绑定
	var createRequest dto.CreateRequest
	if err = ctx.ShouldBindJSON(&createRequest); err != nil {
		slog.Error("参数绑定失败", "error", utils2.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 创建帖子
	postDTO, err := hdl.postSvc.Create(ctx, uid, createRequest.Title, createRequest.Content, createRequest.ContentType)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	// 建立标签
	err = hdl.tagSvc.Bind(ctx, postDTO.ID, createRequest.Tags)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "帖子创建成功", postDTO)
}

// Delete 删除帖子
func (hdl *PostHandler) Delete(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 再拿帖子 pid
	pid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 进行删除
	err = hdl.postSvc.Delete(ctx, pid, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "帖子删除成功", nil)
}

// Update 修改帖子
func (hdl *PostHandler) Update(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 拿帖子 id
	pid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 参数绑定
	var updateRequest dto.UpdateRequest
	err = ctx.ShouldBindJSON(&updateRequest)

	if err != nil {
		slog.Error("参数绑定失败", "error", utils2.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 修改
	err = hdl.postSvc.Update(ctx, pid, uid, updateRequest.Title, updateRequest.Content, updateRequest.Tags)
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
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
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
	ok := hdl.postSvc.Belong(ctx, pid, uid)
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
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil || pageNo < 1 || pageSize > 100 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	total, postDTOs, err := hdl.postSvc.ListByPageAndUid(ctx, uid, pageNo, pageSize)
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
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	err = hdl.postSvc.Like(ctx, pid, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "", nil)
}

func (hdl *PostHandler) Unlike(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	err = hdl.postSvc.Unlike(ctx, pid, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "", nil)
}

func (hdl *PostHandler) IfLike(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	ok, err := hdl.postSvc.IfLike(ctx, pid, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "", ok)
}

func (hdl *PostHandler) Top(ctx *gin.Context) {
	postDTOs, err := hdl.postSvc.Top(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取热门帖子榜单成功", postDTOs)
}

func (hdl *PostHandler) Create(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	// 获取参数并校验
	var createReq comment.CreateRequest
	if err := ctx.ShouldBindJSON(&createReq); err != nil || createReq.ParentID < 0 {
		slog.Error("参数绑定失败", "error", utils2.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 调用 service 层创建评论
	commentDTO, err := hdl.commentSvc.Create(ctx, pid, uid, createReq.ParentID, createReq.ReplyID, createReq.Content)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "评论成功", commentDTO)
}

func (hdl *PostHandler) Delete(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取帖子 id
	_, err = strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 从路由中获取参数 cid
	cid, err := strconv.ParseInt(ctx.Param("cid"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 调用 Service 层
	err = hdl.commentSvc.Delete(ctx, uid, cid)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	// 返回数据
	response.Success(ctx, "评论删除成功", nil)
}

func (hdl *PostHandler) ListByPage(ctx *gin.Context) {
	// 获取参数
	pid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil || pageNo < 1 || pageSize > 100 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	total, commentDTOs, err := hdl.commentSvc.List(ctx, pid, pageNo, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	hasMore := pageNo*pageSize < total

	response.Success(ctx, "获取评论列表成功", gin.H{
		"comments": commentDTOs,
		"total":    total,
		"hasMore":  hasMore,
	})
}

func (hdl *PostHandler) ListReplies(ctx *gin.Context) {
	// 获取参数
	// 从路由中获取 cid 参数
	cid, err := strconv.ParseInt(ctx.Param("cid"), 10, 64)
	if err != nil {
		// 获取帖子详情请求的参数不合法
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 从路由中获取 pid 参数
	_, err = strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		// 获取帖子详情请求的参数不合法
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "3"))
	if err1 != nil || err2 != nil || pageNo < 1 || pageSize > 100 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	total, commentDTOs, err := hdl.commentSvc.ListReplies(ctx, cid, pageNo, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	hasMore := pageNo*pageSize < total

	response.Success(ctx, "获取评论回复列表成功", gin.H{
		"comments": commentDTOs,
		"total":    total,
		"hasMore":  hasMore,
	})
}

func (hdl *PostHandler) CheckAuth(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取要查询评论的 cid
	cid, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 查询是否属于
	ok := hdl.commentSvc.CheckAuth(ctx, cid, uid)
	if !ok {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	response.Success(ctx, "", nil)
}
