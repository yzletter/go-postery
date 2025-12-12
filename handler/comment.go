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

type CommentHandler struct {
	CommentService *service.CommentService
	UserService    *service.UserService
	PostService    *service.PostService
}

func NewCommentHandler(commentService *service.CommentService, userService *service.UserService, postService *service.PostService) *CommentHandler {
	return &CommentHandler{
		CommentService: commentService,
		UserService:    userService,
		PostService:    postService,
	}
}

// Create 新建评论
func (hdl *CommentHandler) Create(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 获取参数并校验
	var comment request.CreateCommentRequest
	if err := ctx.ShouldBind(&comment); err != nil || comment.ParentId < 0 {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	// 调用 service 层创建评论
	commentDTO, err := hdl.CommentService.Create(comment.PostId, int(uid), comment.ParentId, comment.ReplyId, comment.Content)
	if err != nil {
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, commentDTO)
}

func (hdl *CommentHandler) Delete(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 从路由中获取参数 cid
	cid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	// 调用 Service 层
	err = hdl.CommentService.Delete(int(uid), cid)
	if err != nil {
		if err.Error() == "评论不存在" {
			response.ServerError(ctx, "")
		} else if err.Error() == "没有删除权限" {
			response.Unauthorized(ctx, "")
		} else if err.Error() == "删除失败" {
			response.ServerError(ctx, "")
		}
		return
	}

	// 返回数据
	response.Success(ctx, nil)
	return
}

func (hdl *CommentHandler) List(ctx *gin.Context) {
	// 获取参数
	pid, err := strconv.Atoi(ctx.Param("post_id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	comments := hdl.CommentService.List(pid)

	response.Success(ctx, comments)
}

func (hdl *CommentHandler) Belong(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx)
	if err != nil {
		response.Unauthorized(ctx, "请先登录")
		return
	}

	// 获取要查询评论的 cid
	cid, err := strconv.Atoi(ctx.Query("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	// 查询是否属于
	ok := hdl.CommentService.Belong(cid, int(uid))
	if !ok {
		response.Unauthorized(ctx, "")
		return
	}

	response.Success(ctx, nil)
}
