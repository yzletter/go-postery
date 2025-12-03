package dto

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

type PostDTO struct {
	Id        int     `json:"id,omitempty"`
	Title     string  `json:"Title,omitempty"`
	Content   string  `json:"content,omitempty"`
	CreatedAt string  `json:"createdAt,omitempty"`
	Author    UserDTO `json:"author,omitempty"`
}

func ToPostDTO(post model.Post, user model.User) PostDTO {
	return PostDTO{
		Id:        post.Id,
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: post.CreateTime.Format(time.RFC3339),
		Author:    ToUserDTO(user),
	}
}
