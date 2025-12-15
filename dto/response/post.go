package dto

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

type PostDetailDTO struct {
	ID           int64        `json:"id,string"`
	ViewCount    int          `json:"view_count"`
	LikeCount    int          `json:"like_count"`
	CommentCount int          `json:"comment_count"`
	Title        string       `json:"title"`
	Content      string       `json:"content"`
	CreatedAt    string       `json:"createdAt"`
	Author       UserBriefDTO `json:"author"`
	Tags         []string     `json:"tags"`
}

type PostBriefDTO struct {
	ID        int64        `json:"id,string"`
	Title     string       `json:"title"`
	CreatedAt string       `json:"createdAt"`
	Author    UserBriefDTO `json:"author"`
}

func ToPostDetailDTO(post model.Post, user model.User) PostDetailDTO {
	return PostDetailDTO{
		ID:           post.ID,
		Title:        post.Title,
		Content:      post.Content,
		CreatedAt:    post.CreatedAt.Format(time.RFC3339),
		Author:       ToUserBriefDTO(user),
		ViewCount:    post.ViewCount,
		CommentCount: post.CommentCount,
		LikeCount:    post.LikeCount,
		Tags:         nil,
	}
}

func ToPostBriefDTO(post model.Post, user model.User) PostBriefDTO {
	return PostBriefDTO{
		ID:        post.ID,
		Title:     post.Title,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
		Author:    ToUserBriefDTO(user),
	}
}
