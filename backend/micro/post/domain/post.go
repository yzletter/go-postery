package domain

import (
	"strings"
	"time"

	"github.com/yzletter/go-postery/backend/micro/post/model"
)

const postBriefMaxRuneCount = 120

// Post 帖子
type Post struct {
	ID           int64     // 帖子 ID
	User         User      // 作者
	ViewCount    int       // 浏览量
	LikeCount    int       // 点赞数
	CommentCount int       // 评论数
	Status       int       // 状态 1 正常, 2 封禁
	Title        string    // 标题
	Content      string    // 正文
	ContentType  int       // 正文类型 0 普通文本 1 markdown 文本
	Tags         []string  // 标签
	CreatedAt    time.Time // 创建时间
	UpdatedAt    time.Time // 更新时间
}

type PostBrief struct {
	ID        int64
	User      User      // 作者
	Title     string    // 标题
	Abstract  string    // 摘要
	CreatedAt time.Time // 创建时间
}

func (p Post) Briefed() PostBrief {
	return PostBrief{
		ID:        p.ID,
		User:      p.User,
		Title:     p.Title,
		Abstract:  briefContent(p.Content),
		CreatedAt: p.CreatedAt,
	}
}

func briefContent(content string) string {
	content = strings.Join(strings.Fields(content), " ")
	runes := []rune(content)
	if len(runes) <= postBriefMaxRuneCount {
		return content
	}
	return string(runes[:postBriefMaxRuneCount]) + "..."
}

// User 作者
type User struct {
	UserID   int64
	Nickname string
	Avatar   string
}

// ToModelPost domain.Post 转 model.Post
func ToModelPost(post Post) model.Post {
	return model.Post{
		ID:           post.ID,
		UserID:       post.User.UserID,
		ViewCount:    post.ViewCount,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		Status:       post.Status,
		Title:        post.Title,
		Content:      post.Content,
		ContentType:  post.ContentType,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	}
}

// ToModelPosts []domain.Post 转 []*model.Post
func ToModelPosts(posts []Post) []*model.Post {
	res := make([]*model.Post, 0, len(posts))
	for _, post := range posts {
		m := ToModelPost(post)
		res = append(res, &m)
	}
	return res
}

// ToDomainPost model.Post 转 domain.Post
func ToDomainPost(post *model.Post) Post {
	return Post{
		ID:           post.ID,
		User:         User{UserID: post.UserID},
		ViewCount:    post.ViewCount,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		Status:       post.Status,
		Title:        post.Title,
		Content:      post.Content,
		ContentType:  post.ContentType,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	}
}

// ToDomainPosts []*model.Post 转 []domain.Post
func ToDomainPosts(posts []*model.Post) []Post {
	res := make([]Post, 0, len(posts))
	for _, post := range posts {
		res = append(res, ToDomainPost(post))
	}
	return res
}
