package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	database "github.com/yzletter/go-postery/database/gorm"
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
	total, posts := database.GetPostByPage(pageNo, pageSize)
	postsBack := []gin.H{}
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		user := database.GetUserById(post.UserId)
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
	post := database.GetPostByID(pid)
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
	user := database.GetUserById(post.UserId)
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
