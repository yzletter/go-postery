package post

import (
	"time"

	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/model"
)

type DetailDTO struct {
	ID           int64            `json:"id,string"`
	ViewCount    int              `json:"view_count"`
	LikeCount    int              `json:"like_count"`
	CommentCount int              `json:"comment_count"`
	Title        string           `json:"title"`
	Content      string           `json:"content"`
	CreatedAt    string           `json:"createdAt"`
	Author       userdto.BriefDTO `json:"author"`
	Tags         []string         `json:"tags"`
}

type BriefDTO struct {
	ID        int64            `json:"id,string"`
	Title     string           `json:"title"`
	CreatedAt string           `json:"createdAt"`
	Author    userdto.BriefDTO `json:"author"`
}

func ToDetailDTO(post *model.Post, user *model.User) DetailDTO {
	return DetailDTO{
		ID:           post.ID,
		Title:        post.Title,
		Content:      post.Content,
		CreatedAt:    post.CreatedAt.Format(time.RFC3339),
		Author:       userdto.ToBriefDTO(user),
		ViewCount:    post.ViewCount,
		CommentCount: post.CommentCount,
		LikeCount:    post.LikeCount,
		Tags:         nil,
	}
}

func ToBriefDTO(post *model.Post, user *model.User) BriefDTO {
	return BriefDTO{
		ID:        post.ID,
		Title:     post.Title,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
		Author:    userdto.ToBriefDTO(user),
	}
}
