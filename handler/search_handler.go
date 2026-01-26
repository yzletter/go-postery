package handler

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/conf"
	"github.com/yzletter/go-postery/dto/search"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type SearchHandler struct {
	searchSvc service.SearchService
}

func NewSearchHandler(searchSvc service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchSvc: searchSvc,
	}
}

func (hdl *SearchHandler) Search(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	_, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 参数绑定
	var req search.SearchRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	querys := strings.Split(req.Query, " ")
	postDTOs, err := hdl.searchSvc.Search(ctx, querys)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "搜索成功", postDTOs)
}
