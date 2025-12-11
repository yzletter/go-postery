package dto

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

type PostDetailDTO struct {
	Id        int          `json:"id,string"`
	ViewCount int          `json:"view_count"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	CreatedAt string       `json:"createdAt"`
	Author    UserBriefDTO `json:"author"`
}

type PostBriefDTO struct {
	Id        int          `json:"id,string"`
	Title     string       `json:"title"`
	CreatedAt string       `json:"createdAt"`
	Author    UserBriefDTO `json:"author"`
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
