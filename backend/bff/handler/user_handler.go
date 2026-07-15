package handler

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	interactive_grpc "github.com/yzletter/go-postery/api/proto/interactive/v1"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	post_dto "github.com/yzletter/go-postery/backend/bff/dto/post"
	user_dto "github.com/yzletter/go-postery/backend/bff/dto/user"
	"github.com/yzletter/go-postery/backend/bff/errno"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/utils"
	"github.com/yzletter/go-postery/backend/utils/response"
	"google.golang.org/grpc/codes"
)

type UserHandler struct {
	userSvc  manager.UserClient
	postSvc  manager.PostClient
	interSvc manager.InteractiveClient
}

// NewUserHandler 构造函数
func NewUserHandler(userSvc manager.UserClient, postSvc manager.PostClient, interSvc manager.InteractiveClient) *UserHandler {
	return &UserHandler{
		userSvc:  userSvc,
		postSvc:  postSvc,
		interSvc: interSvc,
	}
}

func (hdl *UserHandler) RegisterRouter(engine *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	users := engine.Group("/users")
	users.GET("/:id", hdl.Profile)                    // GET /users/:id									获取个人资料
	users.GET("/:id/posts", hdl.Posts)                // GET /users/:id/posts?pageNo=1&pageSize=10		按页获取用户所发帖子
	users.GET("/top", hdl.Top)                        // GET /users/top 								获取用户榜单
	users.POST("/presign", hdl.GetAvatarURL)          // POST /users/presign 							获取下载头像预签名
	users.POST("/callback", hdl.UploadAvatarCallback) // POST /users/callback 							回调

	authedUsers := users.Group("")
	authedUsers.Use(authMiddleware)

	// 个人模块
	me := authedUsers.Group("/me")
	me.POST("", hdl.ModifyProfile)          // POST 	/users/me										修改个人资料
	me.GET("/upload", hdl.UploadAvatarSign) // GET 	/users/me/upload								获取上传头像签名
	me.GET("/followers", hdl.ListFollowers) // GET 	/users/me/followers?pageNo=1&pageSize=10		按页获取用户粉丝
	me.GET("/followees", hdl.ListFollowees) // GET 	/users/me/followees?pageNo=1&pageSize=10 		按页获取用户关注的人

	// 关注模块
	follow := authedUsers.Group("/:id")
	{
		follow.POST("/follow", hdl.Follow)     // POST 	/users/:id/follow 		关注
		follow.POST("/unfollow", hdl.UnFollow) // Post 	/users/:id/unfollow 	取关
		follow.GET("/follow", hdl.IfFollow)    // GET 	/users/:id/follow 		是否关注
	}
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
				ID:       profileResp.UserID,
				Nickname: profileResp.Nickname,
				Avatar:   profileResp.Avatar,
				Bio:      profileResp.Bio,
			},
		})
	}

	response.Success(ctx, "获取用户发表的帖子成功", gin.H{
		"total":    int64(resp.Count),
		"has_more": pageNo*pageSize < int(resp.Count),
		"posts":    postsBack,
	})
}

// ModifyProfile 修改个人资料
func (hdl *UserHandler) ModifyProfile(ctx *gin.Context) {
	var modifyProfileReq user_dto.ModifyProfileRequest
	// 将请求参数绑定到结构体
	if err := ctx.ShouldBindJSON(&modifyProfileReq); err != nil {
		// 参数绑定失败
		slog.Warn("bind request failed", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	updateReq := &user_grpc.UpdateProfileRequest{
		UserID: uid,
	}
	if modifyProfileReq.Nickname != nil {
		updateReq.Nickname = modifyProfileReq.Nickname
	}
	if modifyProfileReq.Avatar != nil {
		updateReq.Avatar = modifyProfileReq.Avatar
	}
	if modifyProfileReq.Bio != nil {
		updateReq.Bio = modifyProfileReq.Bio
	}
	if modifyProfileReq.Gender != nil {
		gender := int32(*modifyProfileReq.Gender)
		updateReq.Gender = &gender
	}
	if modifyProfileReq.BirthDay != nil {
		updateReq.Birthday = modifyProfileReq.BirthDay
	}
	if modifyProfileReq.Location != nil {
		updateReq.Location = modifyProfileReq.Location
	}
	if modifyProfileReq.Country != nil {
		updateReq.Country = modifyProfileReq.Country
	}

	_, err = hdl.userSvc.UpdateProfile(ctx, updateReq)
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
			codes.NotFound:        errno.ErrUserNotFound,
			codes.AlreadyExists:   errno.ErrNicknameDuplicated,
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
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), []user_dto.TopDTO{})
		return
	}

	res := make([]user_dto.TopDTO, 0, len(resp.ProfileTops))
	for _, u := range resp.ProfileTops {
		top := user_dto.ToTopDTO(u)
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
	if _, err = hdl.interSvc.Follow(ctx, &interactive_grpc.FollowRequest{FollowerID: uid, FolloweeID: id}); err != nil {
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
	if _, err = hdl.interSvc.Unfollow(ctx, &interactive_grpc.FollowRequest{FollowerID: uid, FolloweeID: id}); err != nil {
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

	resp, err := hdl.interSvc.IfFollow(ctx, &interactive_grpc.FollowRequest{FollowerID: uid, FolloweeID: id})
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
	if err1 != nil || err2 != nil || pageNo < 1 || pageSize < 1 || pageSize > 100 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	resp, err := hdl.userSvc.ListFollowersByPage(ctx, &user_grpc.ListFollowRequest{UserID: uid, PageNo: uint32(pageNo), PageSize: uint32(pageSize)})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
		}, errno.ErrServerInternal), gin.H{
			"followers": []user_dto.BriefDTO{},
			"total":     0,
			"hasMore":   false,
		})
		return
	}

	hasMore := pageNo*pageSize < int(resp.Count)

	response.Success(ctx, "获取粉丝列表成功",
		gin.H{
			"followers": user_dto.BriefsToDTO(resp.ProfileBriefs),
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
	if err1 != nil || err2 != nil || pageNo < 1 || pageSize < 1 || pageSize > 100 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	resp, err := hdl.userSvc.ListFolloweesByPage(ctx, &user_grpc.ListFollowRequest{UserID: uid, PageNo: uint32(pageNo), PageSize: uint32(pageSize)})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
		}, errno.ErrServerInternal), gin.H{
			"followers": []user_dto.BriefDTO{},
			"total":     0,
			"hasMore":   false,
		})
		return
	}

	hasMore := pageNo*pageSize < int(resp.Count)

	response.Success(ctx, "获取关注列表成功",
		gin.H{
			"followers": user_dto.BriefsToDTO(resp.ProfileBriefs),
			"total":     resp.Count,
			"hasMore":   hasMore,
		})
}

// GetAvatarURL 预签名
func (hdl *UserHandler) GetAvatarURL(ctx *gin.Context) {
	var req user_dto.GetAvatarURLRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 参数绑定失败
		slog.Warn("bind request failed", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	if !strings.HasPrefix(req.Avatar, "users/avatar/") {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	resp, err := hdl.userSvc.GetAvatarURL(ctx, &user_grpc.GetAvatarURLRequest{ObjectName: req.Avatar})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
		}, errno.ErrServerInternal), nil)
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
		slog.Warn("oss callback signature invalid", "error", err, "method", ctx.Request.Method, "path", ctx.Request.URL.Path, "client_ip", ctx.ClientIP()) // 验签失败
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// Body 绑定结构体
	var uploadCallbackReq user_dto.UploadCallbackRequest
	if err := ctx.ShouldBindJSON(&uploadCallbackReq); err != nil {
		// 参数绑定失败
		slog.Warn("bind request failed", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 验证 Bucket
	if uploadCallbackReq.Bucket != "go-postery" {
		slog.Warn("invalid oss bucket", "bucket", uploadCallbackReq.Bucket)
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
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
			codes.NotFound:        errno.ErrUserNotFound,
		}, errno.ErrServerInternal), nil)
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
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
		}, errno.ErrServerInternal),
			user_dto.OSSSignDTO{
				Response: "",
			})
		return
	}

	response.Success(ctx, "获取签名成功", user_dto.OSSSignDTO{
		Response: resp.Response,
	})
}
