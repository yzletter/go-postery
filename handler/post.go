package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/dto"
	database2 "github.com/yzletter/go-postery/repository/gorm"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
)

type PostHandler struct {
	PostService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{
		PostService: postService,
	}
}

// GetPosts 获取帖子列表
func (postHandler *PostHandler) GetPosts(ctx *gin.Context) {
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

// GetPostDetail 获取帖子详情
func (postHandler *PostHandler) GetPostDetail(ctx *gin.Context) {
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

// CreateNewPost 创建帖子
func (postHandler *PostHandler) CreateNewPost(ctx *gin.Context) {
	// 直接从 ctx 中拿 loginUid
	loginUid := ctx.Value(service.UID_IN_CTX).(int)

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

// DeletePost 删除帖子
func (postHandler *PostHandler) DeletePost(ctx *gin.Context) {
	// 直接从 ctx 中拿 loginUid
	loginUid := ctx.Value(service.UID_IN_CTX).(int)

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

// UpdatePost 修改帖子
func (postHandler *PostHandler) UpdatePost(ctx *gin.Context) {
	// 直接从 ctx 中拿 loginUid
	loginUid := ctx.Value(service.UID_IN_CTX).(int)

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

// PostBelong 查询帖子作者是否为当前登录用户
func (postHandler *PostHandler) PostBelong(ctx *gin.Context) {
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

	// 前面中间件放了 uid 在 ctx, 直接拿 uid
	uid, ok := ctx.Value(service.UID_IN_CTX).(int)
	if !ok {
		// 未登录
		resp := utils.Resp{
			Code: 0,
			Msg:  "帖子不属于当前用户",
			Data: "false",
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}

	// 判断登录用户是否是作者
	post := database2.GetPostByID(pid)
	if post == nil || uid != post.UserId {
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
