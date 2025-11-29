package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/dto"
	"github.com/yzletter/go-postery/middleware"
	"github.com/yzletter/go-postery/middleware/auth"
	database2 "github.com/yzletter/go-postery/repository/gorm"
	"github.com/yzletter/go-postery/repository/redis"
	"github.com/yzletter/go-postery/utils"
)

// GetPostsHandler 获取帖子列表
func GetPostsHandler(ctx *gin.Context) {
	// 从 /posts?pageNo=1&pageSize=2 路由中拿出 pageNo 和 pageSize
	pageNo, err1 := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, err2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err1 != nil || err2 != nil {
		res := utils.Resp{
			Code: 1,
			Msg:  "获取帖子列表请求的参数不合法",
			Data: nil,
		}
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// 获取帖子总数和当前页帖子列表
	total, posts := database2.GetPostByPage(pageNo, pageSize)
	postsBack := []gin.H{}
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		user := database2.GetUserById(post.UserId)
		if user != nil {
			post.UserName = user.Name
		} else {
			slog.Warn("could not get name of user", "uid", post.UserId)
		}

		res := gin.H{
			"id":      post.Id,
			"title":   post.Title,
			"content": post.Content,
			"author": gin.H{
				"id":   post.UserId,
				"name": post.UserName,
			},
			"createdAt": post.ViewTime,
		}
		postsBack = append(postsBack, res)
	}

	// 计算是否还有帖子 = 判断已经加载的帖子数是否小于总帖子数
	hasMore := pageNo*pageSize < total

	resp := utils.Resp{
		Code: 0,
		Msg:  "获取帖子列表成功",
		Data: gin.H{
			"posts":   postsBack,
			"total":   total,
			"hasMore": hasMore,
		},
	}
	ctx.JSON(http.StatusOK, resp)
	return
}

// GetPostDetailHandler 获取帖子详情
func GetPostDetailHandler(ctx *gin.Context) {
	// 从路由中获取 pid 参数
	pid, err := strconv.Atoi(ctx.Param("pid"))
	if err != nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "获取帖子详情失败",
			Data: nil,
		}
		ctx.JSON(http.StatusOK, resp)
	}

	// 根据 pid 查找帖子详情
	post := database2.GetPostByID(pid)
	if post == nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "获取帖子详情失败",
			Data: nil,
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}

	// 获取作者用户名
	user := database2.GetUserById(post.UserId)
	if user != nil {
		post.UserName = user.Name
	} else {
		slog.Warn("could not get name of user", "uid", post.UserId)
	}

	resp := utils.Resp{
		Code: 0,
		Msg:  "获取帖子详情成功",
		Data: gin.H{
			"id":      post.Id,
			"title":   post.Title,
			"content": post.Content,
			"author": gin.H{
				"id":   post.UserId,
				"name": post.UserName,
			},
			"createdAt": post.ViewTime,
		},
	}
	ctx.JSON(http.StatusOK, resp)
}

// CreateNewPostHandler 创建帖子
func CreateNewPostHandler(ctx *gin.Context) {
	// 直接从 ctx 中拿 loginUid
	loginUid := ctx.Value(auth.UID_IN_CTX).(int)

	// 参数绑定
	var createRequest dto.CreateRequest
	err := ctx.ShouldBind(&createRequest)
	if err != nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "创建帖子参数错误",
			Data: nil,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 创建帖子
	pid, err := database2.CreatePost(loginUid, createRequest.Title, createRequest.Content)
	if err != nil {
		// 创建帖子失败
		resp := utils.Resp{
			Code: 1,
			Msg:  "创建帖子失败,请稍后重试",
			Data: nil,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := utils.Resp{
		Code: 0,
		Msg:  "创建帖子成功",
		Data: gin.H{
			"id": pid,
		},
	}
	ctx.JSON(http.StatusOK, resp)
}

// DeletePostHandler 删除帖子
func DeletePostHandler(ctx *gin.Context) {
	// 直接从 ctx 中拿 loginUid
	loginUid := ctx.Value(auth.UID_IN_CTX).(int)

	// 再拿帖子 pid
	pid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || pid == 0 {
		resp := utils.Resp{
			Code: 1,
			Msg:  "帖子 id 获取失败",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 判断登录用户是否是作者
	post := database2.GetPostByID(pid)
	if post == nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "当前帖子不存在",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	} else if loginUid != post.UserId {
		// 无权限删除
		resp := utils.Resp{
			Code: 1,
			Msg:  "无权限删除该帖子",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 进行删除
	err = database2.DeletePost(pid)
	if err != nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "帖子删除失败",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	resp := utils.Resp{
		Code: 0,
		Msg:  "帖子删除成功",
	}
	ctx.JSON(http.StatusOK, resp)
	return
}

// UpdatePostHandler 修改帖子
func UpdatePostHandler(ctx *gin.Context) {
	// 直接从 ctx 中拿 loginUid
	loginUid := ctx.Value(auth.UID_IN_CTX).(int)

	// 参数绑定
	var updateRequest dto.UpdateRequest
	err := ctx.ShouldBind(&updateRequest)
	if err != nil || updateRequest.Id == 0 {
		resp := utils.Resp{
			Code: 1,
			Msg:  "修改帖子参数错误",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 判断登录用户是否是作者
	post := database2.GetPostByID(updateRequest.Id)
	if post == nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "当前帖子不存在",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	} else if loginUid != post.UserId {
		// 无权限删除
		resp := utils.Resp{
			Code: 1,
			Msg:  "无权限修改该帖子",
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// 修改
	err = database2.UpdatePost(updateRequest.Id, updateRequest.Title, updateRequest.Content)
	if err != nil {
		resp := utils.Resp{
			Code: 1,
			Msg:  "修改失败，请稍后重试",
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := utils.Resp{
		Code: 0,
		Msg:  "帖子修改成功",
	}
	ctx.JSON(http.StatusOK, resp)
	return
}

// PostBelongHandler 查询帖子作者是否为当前登录用户
func PostBelongHandler(ctx *gin.Context) {
	// 获取帖子 id
	pid, err := strconv.Atoi(ctx.Query("id"))
	if err != nil {
		resp := utils.Resp{
			Code: 0,
			Msg:  "帖子不属于当前用户",
			Data: "false",
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}

	// 获取登录 uid
	accessToken := middleware.GetTokenFromCookie(ctx, auth.ACCESS_TOKEN_COOKIE_NAME)
	userInfo := middleware.GetUserInfoFromJWT(accessToken)

	slog.Info("Auth", "user", userInfo)

	if userInfo == nil || userInfo.Id == 0 {
		// AccessToken 认证不通过, 尝试通过 RefreshToken  认证
		refreshToken := middleware.GetTokenFromCookie(ctx, auth.REFRESH_TOKEN_COOKIE_NAME)
		result := redis.GoPosteryRedisClient.Get(auth.REFRESH_KEY_PREFIX + refreshToken)
		if result.Err() != nil { // 没拿到 redis 中存的 accessToken
			// RefreshToken 也认证不通过, 没招了, 未登录, 后面不用看了
			slog.Info("Auth", "error", result.Err())

			resp := utils.Resp{
				Code: 0,
				Msg:  "帖子不属于当前用户",
				Data: "false",
			}
			ctx.JSON(http.StatusOK, resp)
			return
		}

		// 如果 redis 能拿到, 重新放到 Cookie 中
		accessToken = result.Val()
		userInfo = middleware.GetUserInfoFromJWT(accessToken)
		if userInfo == nil {
			// 虽然拿到了, 但是有问题 (很小概率)
			resp := utils.Resp{
				Code: 0,
				Msg:  "帖子不属于当前用户",
				Data: "false",
			}
			ctx.JSON(http.StatusOK, resp)
			return
		}

		// 拿到了 AccessToken, 并且一切正常, 放入 Cookie 继续判断
		ctx.SetCookie(auth.ACCESS_TOKEN_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)
	}

	// 判断登录用户是否是作者
	post := database2.GetPostByID(pid)
	if post == nil || userInfo.Id != post.UserId {
		resp := utils.Resp{
			Code: 0,
			Msg:  "帖子不属于当前用户",
			Data: "false",
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}

	// 属于
	resp := utils.Resp{
		Code: 0,
		Msg:  "帖子属于当前用户",
		Data: "true",
	}
	ctx.JSON(http.StatusOK, resp)
	return
}
