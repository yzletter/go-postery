package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/dto/comment"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type CommentHandler struct {
	commentSvc service.CommentService
	userSvc    service.UserService
	postSvc    service.PostService
}

func NewCommentHandler(commentService service.CommentService, userService service.UserService, postService service.PostService) *CommentHandler {
	return &CommentHandler{
		commentSvc: commentService,
		userSvc:    userService,
		postSvc:    postService,
	}
}

// Create 新建评论
func (hdl *CommentHandler) Create(ctx *gin.Context) {
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

	// 获取参数并校验
	var createReq comment.CreateRequest
	if err := ctx.ShouldBindJSON(&createReq); err != nil || createReq.ParentID < 0 {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
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

func (hdl *CommentHandler) Delete(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
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

func (hdl *CommentHandler) ListByPage(ctx *gin.Context) {
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

func (hdl *CommentHandler) ListReplies(ctx *gin.Context) {
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

func (hdl *CommentHandler) CheckAuth(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
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
