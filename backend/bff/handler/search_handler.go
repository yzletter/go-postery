package handler

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	postdto "github.com/yzletter/go-postery/backend/bff/dto/post"
	"github.com/yzletter/go-postery/backend/bff/dto/search"
	"github.com/yzletter/go-postery/backend/bff/errno"
	"github.com/yzletter/go-postery/backend/conf"
	grpcclient "github.com/yzletter/go-postery/backend/grpc/manager"
	utils2 "github.com/yzletter/go-postery/backend/utils"
	"github.com/yzletter/go-postery/backend/utils/response"
)

type SearchHandler struct {
	searchSvc grpcclient.SearchClient
	postSvc   grpcclient.PostClient
	userSvc   grpcclient.UserClient
}

func NewSearchHandler(searchSvc grpcclient.SearchClient, postSvc grpcclient.PostClient, userSvc grpcclient.UserClient) *SearchHandler {
	return &SearchHandler{
		searchSvc: searchSvc,
		postSvc:   postSvc,
		userSvc:   userSvc,
	}
}

func (hdl *SearchHandler) RegisterRouter(engine *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// 搜索模块
	search := engine.Group("/search")
	search.Use(authMiddleware)
	{
		search.POST("", hdl.Search)
	}
}

func (hdl *SearchHandler) Search(ctx *gin.Context) {
	// 校验登录态
	if _, err := utils2.GetUidFromCTX(ctx, conf.UserIDInContext); err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 参数绑定
	var req search.SearchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("参数绑定失败", "error", utils2.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 拆分查询词, 多个查询词默认取交集
	queries := strings.Split(req.Query, " ")

	// 调用 SearchService 获取命中的 PostID
	resp, err := hdl.searchSvc.Search(ctx, &search_grpc.SearchRequest{Queries: queries})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), []postdto.DetailDTO{})
		return
	}

	posts := make([]postdto.DetailDTO, 0, len(resp.DocumentIDs))
	for _, DocID := range resp.DocumentIDs {
		// 解析 PostID
		postID, err := strconv.ParseInt(DocID.DocID, 10, 64)
		if err != nil {
			continue
		}

		// 查询帖子详情
		postDetail, err := hdl.postSvc.GetDetailByID(ctx, &post_grpc.GetDetailByIDRequest{PostID: postID, AddViewCnt: false})
		if err != nil {
			continue
		}

		// 查询作者信息
		user, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: postDetail.UserID})
		if err != nil {
			user = &user_grpc.Profile{}
		}

		// 转 DTO
		posts = append(posts, postdto.ToDetailDTO(postDetail, user))
	}

	response.Success(ctx, "搜索成功", posts)
}
