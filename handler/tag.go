package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/service"
)

type TagHandler struct {
	TagSvc *service.TagService
}

func NewTagHandler(tagSvc *service.TagService) *TagHandler {
	return &TagHandler{TagSvc: tagSvc}
}

func (hdl *TagHandler) Create(ctx *gin.Context) {
	//name := ctx.ShouldBind()
	//err := hdl.TagSvc.Create(name)
	//if err != nil {
	//	if errors.Is(err, repository.ErrUniqueKeyConflict) {
	//		response.Fail(ctx, response.CodeBadRequest, "标签已存在")
	//	}
	//	response.ServerError(ctx, "")
	//}
	//response.Success(ctx, "")
}
