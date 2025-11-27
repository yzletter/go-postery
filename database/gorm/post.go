package database

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yzletter/go-postery/database/model"
	"gorm.io/gorm"
)

// CreatePost 新建帖子
func CreatePost(uid int, title, content string) (int, error) {
	// 模型映射
	now := time.Now()
	post := model.Post{
		UserId:     uid,     // 作者id
		Title:      title,   // 标题
		Content:    content, // 正文
		CreateTime: &now,
		DeleteTime: nil, // 须显示指定为 nil, 写入数据库为 null,
	}

	// 新建数据
	if err := GoPosteryDB.Create(&post).Error; err != nil {
		slog.Error("帖子发布失败", "title", err)
		return 0, errors.New("帖子发布失败")
	}

	return post.Id, nil
}

// DeletePost 根据帖子 id 删除帖子
func DeletePost(pid int) error {
	tx := GoPosteryDB.Model(&model.Post{}).Where("id=? and delete_time is null", pid).Update("delete_time", time.Now())
	if tx.Error != nil {
		// 删除失败
		slog.Error("帖子删除失败", "pid", pid, "error", tx.Error)
		return errors.New("帖子删除失败")
	} else {
		if tx.RowsAffected == 0 {
			return fmt.Errorf("帖子 %d 不存在", pid)
		} else {
			return nil
		}
	}
}

// UpdatePost 修改帖子
func UpdatePost(pid int, title, content string) error {
	tx := GoPosteryDB.Model(&model.Post{}).Where("id=? and delete_time is null", pid).
		Updates(map[string]interface{}{
			"title":   title,
			"content": content,
		})
	if tx.Error != nil {
		//	修改失败
		slog.Error("帖子修改失败", "pid", pid, "error", tx.Error)
		return errors.New("帖子修改失败")
	} else {
		if tx.RowsAffected == 0 {
			slog.Error("帖子不存在", "pid", pid)

			return fmt.Errorf("帖子 %d 不存在", pid)
		} else {
			return nil
		}
	}
}

// GetPostByID 根据帖子 id 获取帖子信息
func GetPostByID(pid int) *model.Post {
	post := &model.Post{
		Id: pid,
	}
	tx := GoPosteryDB.Select("*").Where("delete_time is null").First(post) // find 不会报 ErrNotFound
	if tx.Error != nil {
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) { // 并非未找到, 而是其他错误
			slog.Error("帖子查找失败", "pid", pid, "error", tx.Error)
		}
		return nil
	}

	// 赋值前端用于显示的时间
	post.ViewTime = post.CreateTime.Format("2006-01-02 15:04:05")
	return post
}

// GetPostByPage 翻页查询帖子, 页号从 1 开始, 返回帖子总数和帖子列表
func GetPostByPage(pageNo, pageSize int) (int, []*model.Post) {
	// 获取帖子总数
	var total int64
	tx := GoPosteryDB.Model(&model.Post{}).Where("delete_time is null").Count(&total)
	if tx.Error != nil {
		slog.Error("获取帖子总数失败", "error", tx.Error)
		return 0, nil
	}

	// 获取当前页的帖子
	var posts []*model.Post
	// 已经查询过 pageSize * (pageNo - 1) 条数据, 当前页需要 pageSize 条数据，并按发布时间降序排列
	tx = GoPosteryDB.Model(&model.Post{}).Where("delete_time is null").Order("create_time desc").Limit(pageSize).Offset(pageSize * (pageNo - 1)).Find(&posts)
	if tx.Error != nil {
		slog.Error("获取当前页帖子失败", "pageNo", pageNo, "pageSize", pageSize, "error", tx.Error)
		return 0, nil
	}

	// 赋值前端展示时间
	for _, post := range posts {
		post.ViewTime = post.CreateTime.Format("2006-01-02 15:04:05")
	}

	return int(total), posts
}

// GetPostByUid 根据 uid 获取该用户所发帖子
func GetPostByUid(uid int) []*model.Post {
	// todo
	return nil
}
