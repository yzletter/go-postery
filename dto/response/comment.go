package dto

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

type CommentDTO struct {
	Id        int     `json:"id"`
	PostId    int     `json:"post_id"`
	ParentId  int     `json:"parent_id"`
	Content   string  `json:"content"`
	CreatedAt string  `json:"createdAt"`
	Author    UserDTO `json:"author"`
}

func ToCommentDTO(comment model.Comment, user model.User) CommentDTO {
	return CommentDTO{
		Id:        comment.Id,
		PostId:    comment.PostId,
		ParentId:  comment.ParentId,
		Content:   comment.Content,
		CreatedAt: comment.CreateTime.Format(time.RFC3339),
		Author:    ToUserDTO(user),
	}
}
