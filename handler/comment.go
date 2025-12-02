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
	// 直接拿当前登录用户 uid
	uid := ctx.Value(service.UID_IN_CTX).(int)

	// 获取参数并校验
	var comment request.CreateCommentRequest
	if err := ctx.ShouldBind(&comment); err != nil || comment.ParentId < 0 {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.ParamError(ctx, "")
		return
	}

	// 调用 service 层创建评论
	cid, err := hdl.CommentService.Create(comment.PostId, uid, comment.ParentId, comment.Content)
	if err != nil {
		response.ServerError(ctx, "")
		return
	}

	response.Success(ctx, gin.H{
		"id": cid,
	})
}

func (hdl *CommentHandler) Delete(ctx *gin.Context) {
	// 从 ctx 中拿 uid
	uid := ctx.Value(service.UID_IN_CTX).(int)

	// 从路由中获取参数 cid
	cid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	// 调用 Service 层
	err = hdl.CommentService.Delete(uid, cid)
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

// todo

func (hdl *CommentHandler) List(ctx *gin.Context) {
	// 获取参数
	pid, err := strconv.Atoi(ctx.Param("post_id"))
	if err != nil {
		response.ParamError(ctx, "")
		return
	}

	comments := hdl.CommentService.List(pid)

	commentsBack := make([]gin.H, 0)
	for _, comment := range comments {
		res := gin.H{
			"id":      comment.Id,
			"content": comment.Content,
			"author": gin.H{
				"name": comment.UserName,
			},
			"createdAt": comment.ViewTime,
		}
		commentsBack = append(commentsBack, res)
	}
	response.Success(ctx, commentsBack)
}

func (hdl *CommentHandler) Belong(ctx *gin.Context) {

}
