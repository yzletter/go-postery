package handler

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/conf"
	postdto "github.com/yzletter/go-postery/microservice-backend/bff/dto/post"
	"github.com/yzletter/go-postery/microservice-backend/bff/dto/search"
	"github.com/yzletter/go-postery/microservice-backend/bff/errno"
	grpcclient "github.com/yzletter/go-postery/microservice-backend/bff/grpc/client"
	"github.com/yzletter/go-postery/microservice-backend/bff/utils"
	"github.com/yzletter/go-postery/microservice-backend/bff/utils/response"
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

func (hdl *SearchHandler) Search(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	if _, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext); err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 参数绑定
	var req search.SearchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	querys := strings.Split(req.Query, " ")

	// 进行搜索
	resp, err := hdl.searchSvc.Search(ctx, &search_grpc.SearchRequest{Queries: querys})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), []postdto.DetailDTO{})
		return
	}

	posts := make([]postdto.DetailDTO, 0, len(resp.DocumentIDs))
	for _, DocID := range resp.DocumentIDs {
		// PostID
		postID, err := strconv.ParseInt(DocID.DocID, 10, 64)
		if err != nil {
			continue
		}

		// 查找 Post
		post, err := hdl.postSvc.GetDetailByID(ctx, &post_grpc.GetDetailByIDRequest{PostID: postID, AddViewCnt: false})
		if err != nil {
			continue
		}

		// 查找 User
		user, err := hdl.userSvc.GetProfileById(ctx, &user_grpc.GetProfileByIdRequest{ID: post.UserID})
		if err != nil {
			user = &user_grpc.UserDetail{}
		}

		// 转 DTO
		posts = append(posts, postdto.ToDetailDTO(post, user))
	}

	response.Success(ctx, "搜索成功", posts)
}
