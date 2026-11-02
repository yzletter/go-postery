package post

import (
	"github.com/yzletter/go-postery/backend/api/proto/interactive/v1"
	"github.com/yzletter/go-postery/backend/api/proto/post/v1"
	"github.com/yzletter/go-postery/backend/api/proto/user/v1"
	user_dto "github.com/yzletter/go-postery/backend/micro/bff/dto/user"
)

type DetailDTO struct {
	ID           int64             `json:"id,string"`
	ViewCount    int               `json:"view_count"`
	LikeCount    int               `json:"like_count"`
	CommentCount int               `json:"comment_count"`
	Title        string            `json:"title"`
	Content      string            `json:"content"`
	ContentType  int               `json:"content_type"`
	CreatedAt    string            `json:"created_at"`
	Author       user_dto.BriefDTO `json:"author"`
	Tags         []string          `json:"tags"`
}

type BriefDTO struct {
	ID        int64             `json:"id,string"`
	Title     string            `json:"title"`
	Abstract  string            `json:"abstract"`
	CreatedAt string            `json:"created_at"`
	Author    user_dto.BriefDTO `json:"author"`
}

type TopDTO struct {
	ID    int64   `json:"id,string"`
	Title string  `json:"title"`
	Score float64 `json:"score"`
}

func ToDetailDTO(post *post_grpc.PostDetail, user *user_grpc.Profile) DetailDTO {
	return DetailDTO{
		ID:           post.ID,
		Title:        post.Title,
		Content:      post.Content,
		CreatedAt:    post.CreatedAt,
		Author:       user_dto.ToBriefDTO(user),
		ContentType:  int(post.ContentType),
		ViewCount:    int(post.ViewCount),
		CommentCount: int(post.CommentCount),
		LikeCount:    int(post.LikeCount),
		Tags:         post.Tags,
	}
}

func ToBriefDTO(post *post_grpc.PostBrief, user *user_grpc.Profile) BriefDTO {
	return BriefDTO{
		ID:        post.ID,
		Title:     post.Title,
		Abstract:  post.Abstract,
		CreatedAt: post.CreatedAt,
		Author:    user_dto.ToBriefDTO(user),
	}
}

func ToTopDTO(post *post_grpc.TopPost) TopDTO {
	return TopDTO{
		ID:    post.ID,
		Title: post.Title,
		Score: float64(post.Score),
	}
}

type CommentDTO struct {
	ID        int64             `json:"id,string"`
	PostID    int64             `json:"post_id,string"`
	ParentID  int64             `json:"parent_id,string"`
	ReplyID   int64             `json:"reply_id,string"`
	Content   string            `json:"content"`
	CreatedAt string            `json:"created_at"`
	Author    user_dto.BriefDTO `json:"author"`
}

func ToInteractiveCommentDTO(comment *interactive_grpc.InteractiveComment, user *user_grpc.Profile) CommentDTO {
	return CommentDTO{
		ID:        comment.ID,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		ReplyID:   comment.ReplyID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		Author:    user_dto.ToBriefDTO(user),
	}
}
