package dto

import (
	"time"

	"github.com/yzletter/go-postery/model"
)

type CommentDTO struct {
	Id        int          `json:"id,string"`
	PostId    int          `json:"post_id,string"`
	ParentId  int          `json:"parent_id,string"`
	ReplyId   int          `json:"reply_id,string"`
	Content   string       `json:"content"`
	CreatedAt string       `json:"createdAt"`
	Author    UserBriefDTO `json:"author"`
}

func ToCommentDTO(comment model.Comment, user model.User) CommentDTO {
	return CommentDTO{
		Id:        comment.ID,
		PostId:    comment.PostID,
		ParentId:  comment.ParentID,
		ReplyId:   comment.ReplyID,
		Content:   comment.Content,
		CreatedAt: comment.CreateTime.Format(time.RFC3339),
		Author:    ToUserBriefDTO(user),
	}
}
