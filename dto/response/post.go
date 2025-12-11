package dto

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

type PostDetailDTO struct {
	Id        int          `json:"id,omitempty,string"`
	Title     string       `json:"title,omitempty"`
	Content   string       `json:"content,omitempty"`
	CreatedAt string       `json:"createdAt,omitempty"`
	Author    UserBriefDTO `json:"author,omitempty"`
	ViewCount int          `json:"view_count,omitempty"`
}

type PostBriefDTO struct {
	Id        int          `json:"id,omitempty,string"`
	Title     string       `json:"title,omitempty"`
	CreatedAt string       `json:"createdAt,omitempty"`
	Author    UserBriefDTO `json:"author,omitempty"`
}

func ToPostDetailDTO(post model.Post, user model.User) PostDetailDTO {
	return PostDetailDTO{
		Id:        post.Id,
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: post.CreateTime.Format(time.RFC3339),
		Author:    ToUserBriefDTO(user),
		ViewCount: post.ViewCount,
	}
}

func ToPostBriefDTO(post model.Post, user model.User) PostBriefDTO {
	return PostBriefDTO{
		Id:        post.Id,
		Title:     post.Title,
		CreatedAt: post.CreateTime.Format(time.RFC3339),
		Author:    ToUserBriefDTO(user),
	}
}
