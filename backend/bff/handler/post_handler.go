package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	interactive_grpc "github.com/yzletter/go-postery/api/proto/interactive/v1"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/bff/dto/post"
	"github.com/yzletter/go-postery/backend/bff/errno"
	"github.com/yzletter/go-postery/backend/conf"
	grpcclient "github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/utils"
	"github.com/yzletter/go-postery/backend/utils/response"
	"google.golang.org/grpc/codes"
)

type PostHandler struct {
	postSvc  grpcclient.PostClient
	userSvc  grpcclient.UserClient
	interSvc grpcclient.InteractiveClient
}

func NewPostHandler(postSvc grpcclient.PostClient, userSvc grpcclient.UserClient, interSvc grpcclient.InteractiveClient) *PostHandler {
	return &PostHandler{
		postSvc:  postSvc,
		userSvc:  userSvc,
		interSvc: interSvc,
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
	resp, err := hdl.postSvc.ListByPage(ctx, &post_grpc.ListByPageRequest{
		PageNo:   uint32(pageNo),
		PageSize: uint32(pageSize),
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrPostNotFound,
		}, errno.ErrServerInternal), gin.H{
			"posts":   []post.DetailDTO{},
			"total":   0,
			"hasMore": false,
		})
		return
	}

	userDetails := make([]*user_grpc.Profile, len(resp.PostDetails)) // 用户切片
	for k := range resp.PostDetails {
		userDetail, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: resp.PostDetails[k].UserID})
		if err != nil {
			userDetail = &user_grpc.Profile{}
		}
		userDetails[k] = userDetail
	}

	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := pageNo*pageSize < int(resp.Count)

	posts := make([]post.DetailDTO, 0, len(resp.PostDetails))
	for k := range resp.PostDetails {
		posts = append(posts, post.ToDetailDTO(resp.PostDetails[k], userDetails[k]))
	}

	// 返回
	response.Success(ctx, "获取帖子列表成功", gin.H{
		"posts":   posts,
		"total":   resp.Count,
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
	resp, err := hdl.postSvc.ListByPageAndTag(ctx, &post_grpc.ListByPageAndTagRequest{Tag: name, PageNo: uint32(pageNo), PageSize: uint32(pageSize)})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrPostNotFound,
		}, errno.ErrServerInternal), gin.H{
			"posts":   []post.DetailDTO{},
			"total":   0,
			"hasMore": false,
		})
		return
	}

	userDetails := make([]*user_grpc.Profile, len(resp.PostDetails)) // 用户切片
	for k := range resp.PostDetails {
		userDetail, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: resp.PostDetails[k].UserID})
		if err != nil {
			userDetail = &user_grpc.Profile{}
		}
		userDetails[k] = userDetail
	}

	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := pageNo*pageSize < int(resp.Count)

	posts := make([]post.DetailDTO, 0, len(resp.PostDetails))
	for k := range resp.PostDetails {
		posts = append(posts, post.ToDetailDTO(resp.PostDetails[k], userDetails[k]))
	}

	// 返回
	response.Success(ctx, "获取帖子列表成功", gin.H{
		"posts":   posts,
		"total":   resp.Count,
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
	postDetail, err := hdl.postSvc.GetDetailByID(ctx, &post_grpc.GetDetailByIDRequest{PostID: pid, AddViewCnt: true})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrPostNotFound,
		}, errno.ErrServerInternal), post.DetailDTO{})
		return
	}

	// 查询用户
	userDetail, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: postDetail.UserID})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
			codes.NotFound:        errno.ErrUserNotFound,
		}, errno.ErrServerInternal), post.DetailDTO{})
		return
	}

	response.Success(ctx, "获取帖子详情成功", post.ToDetailDTO(postDetail, userDetail))
}

// CreatePost 创建帖子
func (hdl *PostHandler) CreatePost(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 参数绑定
	var createRequest post.CreatePostRequest
	if err = ctx.ShouldBindJSON(&createRequest); err != nil {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 创建帖子
	postDetail, err := hdl.postSvc.Create(ctx, &post_grpc.CreatePostRequest{
		UserID:      uid,
		Title:       createRequest.Title,
		Content:     createRequest.Content,
		ContentType: uint32(createRequest.ContentType),
		Tags:        createRequest.Tags,
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), post.DetailDTO{})
		return
	}

	// 查询用户
	userDetail, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: postDetail.UserID})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
			codes.NotFound:        errno.ErrUserNotFound,
		}, errno.ErrServerInternal), post.DetailDTO{})
		return
	}

	response.Success(ctx, "帖子创建成功", post.ToDetailDTO(postDetail, userDetail))
}

// DeletePost 删除帖子
func (hdl *PostHandler) DeletePost(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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
	_, err = hdl.postSvc.Delete(ctx, &post_grpc.PostCommonRequest{PostID: pid, UserID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound:        errno.ErrPostNotFound,
			codes.Unauthenticated: errno.ErrUnauthorized,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "帖子删除成功", nil)
}

// Update 修改帖子
func (hdl *PostHandler) Update(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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
	var updateRequest post.UpdateRequest
	err = ctx.ShouldBindJSON(&updateRequest)

	if err != nil {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 修改
	if _, err = hdl.postSvc.Update(ctx, &post_grpc.UpdateRequest{
		PostID:  pid,
		UserID:  uid,
		Title:   updateRequest.Title,
		Content: updateRequest.Content,
		Tags:    updateRequest.Tags,
	}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound:        errno.ErrPostNotFound,
			codes.Unauthenticated: errno.ErrUnauthorized,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "帖子更新成功", nil)
	return
}

// Belong 查询帖子作者是否为当前登录用户
func (hdl *PostHandler) Belong(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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
	ok, err := hdl.postSvc.Belong(ctx, &post_grpc.PostCommonRequest{PostID: pid, UserID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrPostNotFound,
		}, errno.ErrServerInternal), gin.H{})
		return
	} else if !ok.Result {
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

	resp, err := hdl.postSvc.ListByPageAndUid(ctx, &post_grpc.ListByPageAndUidRequest{
		UserID:   uid,
		PageNo:   uint32(pageNo),
		PageSize: uint32(pageSize),
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrPostNotFound,
		}, errno.ErrServerInternal), gin.H{
			"posts":   []post.BriefDTO{},
			"total":   0,
			"hasMore": false,
		})
		return
	}

	userDetails := make([]*user_grpc.Profile, len(resp.PostBriefs)) // 用户切片
	for k := range resp.PostBriefs {
		userDetail, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: resp.PostBriefs[k].UserID})
		if err != nil {
			userDetail = &user_grpc.Profile{}
		}
		userDetails[k] = userDetail
	}

	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := pageNo*pageSize < int(resp.Count)

	posts := make([]post.BriefDTO, 0, len(resp.PostBriefs))
	for k := range resp.PostBriefs {
		posts = append(posts, post.ToBriefDTO(resp.PostBriefs[k], userDetails[k]))
	}

	// 返回
	response.Success(ctx, "获取帖子列表成功", gin.H{
		"posts":   posts,
		"total":   resp.Count,
		"hasMore": hasMore,
	})
}

func (hdl *PostHandler) Like(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	if _, err = hdl.interSvc.Like(ctx, &interactive_grpc.LikeRequest{PostID: pid, UserID: uid}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound:      errno.ErrPostNotFound,
			codes.AlreadyExists: errno.ErrDuplicatedLike,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "", nil)
}

func (hdl *PostHandler) Unlike(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	if _, err = hdl.interSvc.Unlike(ctx, &interactive_grpc.LikeRequest{PostID: pid, UserID: uid}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound:      errno.ErrPostNotFound,
			codes.AlreadyExists: errno.ErrDuplicatedUnLike,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "", nil)
}

func (hdl *PostHandler) IfLike(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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

	ok, err := hdl.interSvc.CheckLike(ctx, &interactive_grpc.LikeRequest{PostID: pid, UserID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), false)
		return
	}

	response.Success(ctx, "", ok.Result)
}

func (hdl *PostHandler) Top(ctx *gin.Context) {
	resp, err := hdl.postSvc.Top(ctx, &post_grpc.PostEmptyRequest{})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), []post.TopDTO{})
		return
	}

	res := make([]post.TopDTO, 0, len(resp.TopPosts))
	for _, topPost := range resp.TopPosts {
		res = append(res, post.ToTopDTO(topPost))
	}

	response.Success(ctx, "获取热门帖子榜单成功", res)
}

func (hdl *PostHandler) CreateComment(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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
	var createReq post.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&createReq); err != nil || createReq.ParentID < 0 {
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 调用 interactive service 创建评论
	commentDTO, err := hdl.interSvc.Comment(ctx, &interactive_grpc.CreateCommentRequest{
		PostID:   pid,
		ParentID: createReq.ParentID,
		ReplyID:  createReq.ReplyID,
		UserID:   uid,
		Content:  createReq.Content,
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrPostNotFound,
		}, errno.ErrServerInternal), post.CommentDTO{})
		return
	}

	response.Success(ctx, "评论成功", commentDTO)
}

func (hdl *PostHandler) DeleteComment(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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
	_, err = hdl.interSvc.DelComment(ctx, &interactive_grpc.DeleteCommentRequest{
		UserID:    uid,
		CommentID: cid,
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound:        errno.ErrCommentNotFound,
			codes.Unauthenticated: errno.ErrUnauthorized,
		}, errno.ErrServerInternal), gin.H{})
		return
	}
	// 返回数据
	response.Success(ctx, "评论删除成功", nil)
}

func (hdl *PostHandler) ListCommentByPage(ctx *gin.Context) {
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

	resp, err := hdl.interSvc.ListCommentByPage(ctx, &interactive_grpc.ListCommentByPageRequest{
		PostID:   pid,
		PageNo:   uint32(pageNo),
		PageSize: uint32(pageSize),
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrCommentNotFound,
		}, errno.ErrServerInternal), gin.H{
			"comments": []post.CommentDTO{},
			"total":    0,
			"hasMore":  false,
		})
		return
	}

	hasMore := pageNo*pageSize < int(resp.Count)

	userDetails := make([]*user_grpc.Profile, len(resp.Comments)) // 用户切片
	for k := range resp.Comments {
		userDetail, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: resp.Comments[k].UserID})
		if err != nil {
			userDetail = &user_grpc.Profile{}
		}
		userDetails[k] = userDetail
	}

	comments := make([]post.CommentDTO, 0, len(resp.Comments))
	for k := range resp.Comments {
		comments = append(comments, post.ToInteractiveCommentDTO(resp.Comments[k], userDetails[k]))
	}

	response.Success(ctx, "获取评论列表成功", gin.H{
		"comments": comments,
		"total":    resp.Count,
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

	resp, err := hdl.interSvc.ListRepliesByPage(ctx, &interactive_grpc.ListReplyByPageRequest{
		CommentID: cid,
		PageNo:    uint32(pageNo),
		PageSize:  uint32(pageSize),
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrCommentNotFound,
		}, errno.ErrServerInternal), gin.H{
			"comments": []post.CommentDTO{},
			"total":    0,
			"hasMore":  false,
		})
		return
	}

	hasMore := pageNo*pageSize < int(resp.Count)

	userDetails := make([]*user_grpc.Profile, len(resp.Comments)) // 用户切片
	for k := range resp.Comments {
		userDetail, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: resp.Comments[k].UserID})
		if err != nil {
			userDetail = &user_grpc.Profile{}
		}
		userDetails[k] = userDetail
	}

	comments := make([]post.CommentDTO, 0, len(resp.Comments))
	for k := range resp.Comments {
		comments = append(comments, post.ToInteractiveCommentDTO(resp.Comments[k], userDetails[k]))
	}

	response.Success(ctx, "获取评论回复列表成功", gin.H{
		"comments": comments,
		"total":    resp.Count,
		"hasMore":  hasMore,
	})
}

func (hdl *PostHandler) CheckCommentDeleteAuth(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
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
	resp, err := hdl.interSvc.CheckCommentDelAuth(ctx, &interactive_grpc.CommentIDUserIDRequest{UserID: uid, CommentID: cid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrCommentNotFound,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	if !resp.Result {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	response.Success(ctx, "", nil)
}
