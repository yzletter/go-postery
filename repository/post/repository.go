package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yzletter/go-postery/infra/snowflake"
	"github.com/yzletter/go-postery/model"
	"gorm.io/gorm"
)

type GormPostRepository struct {
	db *gorm.DB
}

func NewGormPostRepository(db *gorm.DB) *GormPostRepository {
	return &GormPostRepository{
		db: db,
	}
}

// Create 新建帖子
func (repo *GormPostRepository) Create(uid int, title, content string) (model.Post, error) {
	// 模型映射
	now := time.Now()
	post := model.Post{
		Id:         snowflake.NextID(),
		UserId:     uid,     // 作者id
		Title:      title,   // 标题
		Content:    content, // 正文
		CreateTime: &now,
		DeleteTime: nil, // 须显示指定为 nil, 写入数据库为 null,
	}

	// 新建数据
	if err := repo.db.Create(&post).Error; err != nil {
		slog.Error("帖子发布失败", "title", err)
		return model.Post{}, errors.New("帖子发布失败")
	}

	return post, nil
}

// Delete 根据帖子 id 删除帖子
func (repo *GormPostRepository) Delete(pid int) error {
	tx := repo.db.Model(&model.Post{}).Where("id=? and delete_time is null", pid).Update("delete_time", time.Now())
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

// Update 修改帖子
func (repo *GormPostRepository) Update(pid int, title, content string) error {
	tx := repo.db.Model(&model.Post{}).Where("id=? and delete_time is null", pid)

	var count int64
	tx.Count(&count)

	if count == 0 {
		slog.Error("帖子不存在", "pid", pid)
		return fmt.Errorf("帖子 %d 不存在", pid)
	}

	// 修改
	tx.Updates(map[string]interface{}{
		"title":   title,
		"content": content,
	})
	if tx.Error != nil {
		//	修改失败
		slog.Error("帖子修改失败", "pid", pid, "error", tx.Error)
		return errors.New("帖子修改失败")
	}
	return nil
}

// GetByID 根据帖子 id 获取帖子信息
func (repo *GormPostRepository) GetByID(pid int) (bool, model.Post) {
	post := model.Post{
		Id: pid,
	}
	tx := repo.db.Select("*").Where("delete_time is null").First(&post) // find 不会报 ErrNotFound
	if tx.Error != nil {
		if !errors.Is(tx.Error, gorm.ErrRecordNotFound) { // 并非未找到, 而是其他错误
			slog.Error("帖子查找失败", "pid", pid, "error", tx.Error)
		}
		return false, model.Post{}
	}

	return true, post
}

// GetByPage 翻页查询帖子, 页号从 1 开始, 返回帖子总数和帖子列表
func (repo *GormPostRepository) GetByPage(pageNo, pageSize int) (int, []model.Post) {
	// 获取帖子总数
	var total int64
	tx := repo.db.Model(&model.Post{}).Where("delete_time is null").Count(&total)
	if tx.Error != nil {
		slog.Error("获取帖子总数失败", "error", tx.Error)
		return 0, nil
	}

	// 获取当前页的帖子
	var posts []model.Post
	// 已经查询过 pageSize * (pageNo - 1) 条数据, 当前页需要 pageSize 条数据，并按发布时间降序排列
	tx = repo.db.Model(&model.Post{}).Where("delete_time is null").Order("create_time desc").Limit(pageSize).Offset(pageSize * (pageNo - 1)).Find(&posts)
	if tx.Error != nil {
		slog.Error("获取当前页帖子失败", "pageNo", pageNo, "pageSize", pageSize, "error", tx.Error)
		return 0, nil
	}

	return int(total), posts
}

// GetByUid 根据 uid 获取该用户所发帖子
func (repo *GormPostRepository) GetByUid(uid int) []model.Post {
	var posts []model.Post
	// 查找五条由 uid 所发的帖子
	tx := repo.db.Model(&model.Post{}).Where("user_id = ? and delete_time is null", uid).Order("create_time desc").Limit(5).Find(&posts)

	if tx.Error != nil {
		return nil
	}
	return posts
}
