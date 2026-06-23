package handler

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	post_dto "github.com/yzletter/go-postery/backend/micro/bff/dto/post"
	user_dto "github.com/yzletter/go-postery/backend/micro/bff/dto/user"
	"github.com/yzletter/go-postery/backend/micro/bff/errno"
	"github.com/yzletter/go-postery/backend/utils"
	"github.com/yzletter/go-postery/backend/utils/response"
	"google.golang.org/grpc/codes"
)

type UserHandler struct {
	userSvc manager.UserClient
	postSvc manager.PostClient
}

// NewUserHandler 构造函数
func NewUserHandler(userSvc manager.UserClient, postSvc manager.PostClient) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		postSvc: postSvc,
	}
}

// RegisterPrivateRouter 注册私人路由
func (hdl *UserHandler) RegisterPrivateRouter(engine gin.IRouter) {
	// 个人模块
	me := engine.Group("/me")
	me.POST("", hdl.ModifyProfile)          // POST 	/users/authed/me										修改个人资料
	me.GET("/upload", hdl.UploadAvatarSign) // GET 	/users/authed/me/upload									获取上传头像签名
	me.GET("/followers", hdl.ListFollowers) // GET 	/users/authed/me/followers?pageNo=1&pageSize=10			按页获取用户粉丝
	me.GET("/followees", hdl.ListFollowees) // GET 	/users/authed/me/followees?pageNo=1&pageSize=10 		按页获取用户关注的人

	// 关注模块
	follow := engine.Group("/:id")
	{
		follow.POST("/follow", hdl.Follow)     // POST 	/users/authed/:id/follow 		关注
		follow.POST("/unfollow", hdl.UnFollow) // Post 	/users/authed/:id/unfollow 		取关
		follow.GET("/follow", hdl.IfFollow)    // GET 	/users/authed/:id/follow 		是否关注
	}
}

// RegisterPublicRouter 注册公共路由
func (hdl *UserHandler) RegisterPublicRouter(engine gin.IRouter) {
	engine.GET("/:id", hdl.Profile)                    // GET /users/:id								获取个人资料
	engine.GET("/:id/posts", hdl.Posts)                // GET /users/:id/posts?pageNo=1&pageSize=10		按页获取用户所发帖子
	engine.GET("/top", hdl.Top)                        // GET /users/top 								获取用户榜单
	engine.POST("/presign", hdl.GetAvatarURL)          // POST /users/presign 							获取下载头像预签名
	engine.POST("/callback", hdl.UploadAvatarCallback) // POST /users/callback 							回调
}

// Profile 获取用户资料
func (hdl *UserHandler) Profile(ctx *gin.Context) {
	uid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	profile, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err,
			map[codes.Code]*errno.Error{
				codes.InvalidArgument: errno.ErrInvalidParam,
				codes.NotFound:        errno.ErrUserNotFound,
			},
			errno.ErrServerInternal), user_dto.DetailDTO{})
		return
	}

	response.Success(ctx, "获取个人资料成功", user_dto.ToDetailDTO(profile))
}

// Posts 按页获取用户发布的帖子
func (hdl *UserHandler) Posts(ctx *gin.Context) {
	// 从 URL 中获取用户 ID
	uid, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 从 URL 中获取 pageNo 和 pageSize
	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 检查用户是否存在
	profileResp, err := hdl.userSvc.GetProfile(ctx, &user_grpc.GetProfileByIdRequest{ID: uid})
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 获取用户发布的帖子
	resp, err := hdl.postSvc.ListByPageAndUid(ctx, &post_grpc.ListByPageAndUidRequest{
		UserID:   uid,
		PageNo:   uint32(pageNo),
		PageSize: uint32(pageSize),
	})

	if err != nil {
		response.Error(ctx, errno.ErrServerInternal)
		return
	}

	postsBack := make([]post_dto.BriefDTO, 0, len(resp.PostBriefs))
	for _, post := range resp.PostBriefs {
		postsBack = append(postsBack, post_dto.BriefDTO{
			ID:        post.ID,
			Title:     post.Title,
			CreatedAt: post.CreatedAt,
			Author: user_dto.BriefDTO{
				ID:       profileResp.ID,
				Nickname: profileResp.Nickname,
				Avatar:   profileResp.Avatar,
			},
		})
	}

	httpBack := user_dto.PostsResponse{
		Total:   int64(resp.Count),
		HasMore: pageNo*pageSize < int(resp.Count),
		Posts:   postsBack,
	}
	response.Success(ctx, "获取用户发表的帖子成功", httpBack)
}

// ModifyProfile 修改个人资料
func (hdl *UserHandler) ModifyProfile(ctx *gin.Context) {
	var modifyProfileReq user.ModifyProfileRequest
	// 将请求参数绑定到结构体
	if err := ctx.ShouldBindJSON(&modifyProfileReq); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	_, err = hdl.userSvc.UpdateProfile(ctx, &user_grpc.UpdateProfileRequest{
		ID:       uid,
		Nickname: modifyProfileReq.Nickname,
		Avatar:   modifyProfileReq.Avatar,
		Bio:      modifyProfileReq.Bio,
		Gender:   uint32(modifyProfileReq.Gender),
		Birthday: modifyProfileReq.BirthDay,
		Location: modifyProfileReq.Location,
		Country:  modifyProfileReq.Country,
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
			codes.NotFound:        errno.ErrUserNotFound,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	// 默认情况下也返回200
	response.Success(ctx, "修改个人资料成功", nil)
}

// Top 获取热门推荐用户
func (hdl *UserHandler) Top(ctx *gin.Context) {
	resp, err := hdl.userSvc.Top(ctx, &user_grpc.TopRequest{})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), []user.TopDTO{})
		return
	}

	res := make([]user.TopDTO, 0, len(resp.TopUsers))
	for _, u := range resp.TopUsers {
		top := user.ToTopDTO(u)
		res = append(res, top)
	}

	response.Success(ctx, "获取推荐关注成功", res)
}

func (hdl *UserHandler) Follow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取对方 id
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	if uid == id {
		response.Error(ctx, errno.ErrFollowYourself)
		return
	}

	// 关注
	if _, err = hdl.userSvc.Follow(ctx, &user_grpc.FollowCommonRequest{FollowerID: uid, FolloweeID: id}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.AlreadyExists: errno.ErrDuplicatedFollow,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "关注成功", nil)
}

func (hdl *UserHandler) UnFollow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取对方 id
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	if uid == id {
		response.Error(ctx, errno.ErrFollowYourself)
		return
	}

	// 取消关注
	if _, err = hdl.userSvc.UnFollow(ctx, &user_grpc.FollowCommonRequest{FollowerID: uid, FolloweeID: id}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.AlreadyExists: errno.ErrDuplicatedUnFollow,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "取消关注成功", nil)
}

func (hdl *UserHandler) IfFollow(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	// 获取对方 id
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	if uid == id {
		response.Error(ctx, errno.ErrFollowYourself)
		return
	}

	resp, err := hdl.userSvc.IfFollow(ctx, &user_grpc.FollowCommonRequest{FollowerID: uid, FolloweeID: id})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), false)
		return
	}

	response.Success(ctx, "获取关注关系成功", resp.Result)
}

// ListFollowers 返回关注我的人
func (hdl *UserHandler) ListFollowers(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil || pageNo < 1 || pageSize > 100 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	resp, err := hdl.userSvc.ListFollowersByPage(ctx, &user_grpc.ListFollowRequest{UserID: uid, PageNo: uint32(pageNo), PageSize: uint32(pageSize)})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), gin.H{
			"followers": []user.BriefDTO{},
			"total":     0,
			"hasMore":   false,
		})
		return
	}

	hasMore := pageNo*pageSize < int(resp.Count)

	response.Success(ctx, "获取粉丝列表成功",
		gin.H{
			"followers": user.BriefsToDTO(resp.UserBriefs),
			"total":     resp.Count,
			"hasMore":   hasMore,
		})
}

// ListFollowees 返回我关注的人
func (hdl *UserHandler) ListFollowees(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil || pageNo < 1 || pageSize > 100 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	resp, err := hdl.userSvc.ListFolloweesByPage(ctx, &user_grpc.ListFollowRequest{UserID: uid, PageNo: uint32(pageNo), PageSize: uint32(pageSize)})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), gin.H{
			"followers": []user.BriefDTO{},
			"total":     0,
			"hasMore":   false,
		})
		return
	}

	hasMore := pageNo*pageSize < int(resp.Count)

	response.Success(ctx, "获取关注列表成功",
		gin.H{
			"followers": user.BriefsToDTO(resp.UserBriefs),
			"total":     resp.Count,
			"hasMore":   hasMore,
		})
}

// GetAvatarURL 预签名
func (hdl *UserHandler) GetAvatarURL(ctx *gin.Context) {
	var req user.GetAvatarURLRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	if !strings.HasPrefix(req.Avatar, "users/avatar/") {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	resp, err := hdl.userSvc.GetAvatarURL(ctx, &user_grpc.GetAvatarURLRequest{ObjectName: req.Avatar})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), nil)
		return
	}

	response.Success(ctx, "获取预签名 URL", gin.H{
		"url": resp.URL,
	})
}

// UploadAvatarCallback 回调
func (hdl *UserHandler) UploadAvatarCallback(ctx *gin.Context) {
	// 对请求进行验签
	if ok, err := utils.VerifyOSS(ctx.Request); !ok || err != nil {
		slog.Error("OSS Callback Invaild Request", "error", err, "request", ctx.Request) // 验签失败
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// Body 绑定结构体
	var uploadCallbackReq user.UploadCallbackRequest
	if err := ctx.ShouldBindJSON(&uploadCallbackReq); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 验证 Bucket
	if uploadCallbackReq.Bucket != "go-postery" {
		slog.Error("Invaild Bucket")
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 获取 uid 和 objectName
	// users/avatar/uid/filename
	segments := strings.Split(uploadCallbackReq.Object, "/")
	if len(segments) != 4 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	uid, err := strconv.ParseInt(segments[2], 10, 64)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	_, err = hdl.userSvc.UploadAvatarCallback(ctx, &user_grpc.UploadAvatarCallbackRequest{UserID: uid, ObjectName: uploadCallbackReq.Object})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), nil)
		return
	}

	response.Success(ctx, "回调成功", nil)
}

// UploadAvatarSign 上传头像
func (hdl *UserHandler) UploadAvatarSign(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	resp, err := hdl.userSvc.UploadAvatarSign(ctx, &user_grpc.UploadAvatarSignRequest{UserID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal),
			user.OSSSignDTO{
				Response: "",
			})
		return
	}

	response.Success(ctx, "获取签名成功", user.OSSSignDTO{
		Response: resp.Response,
	})
}
