package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/dto/request"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type CommentHandler struct {
	CommentService *service.CommentService
	UserService    *service.UserService
}

func NewCommentHandler(commentService *service.CommentService, userService *service.UserService) *CommentHandler {
	return &CommentHandler{
		CommentService: commentService,
		UserService:    userService,
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
