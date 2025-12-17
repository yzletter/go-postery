package comment

import (
	"time"

	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/model"
)

type DTO struct {
	ID        int64            `json:"id,string"`
	PostID    int64            `json:"post_id,string"`
	ParentID  int64            `json:"parent_id,string"`
	ReplyID   int64            `json:"reply_id,string"`
	Content   string           `json:"content"`
	CreatedAt string           `json:"created_at"`
	Author    userdto.BriefDTO `json:"author"`
}

func ToDTO(comment *model.Comment, user *model.User) DTO {
	return DTO{
		ID:        comment.ID,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		ReplyID:   comment.ReplyID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		Author:    userdto.ToBriefDTO(user),
	}
}
