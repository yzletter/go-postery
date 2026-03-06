package dto

import (
	"time"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	model2 "github.com/yzletter/go-postery/microservice-backend/post/model"
)

// ToPostDetail model.Post 转 post_grpc.PostDetail
func ToPostDetail(post *model2.Post, tags []string) *post_grpc.PostDetail {
	return &post_grpc.PostDetail{
		ID:           post.ID,
		ViewCount:    uint64(post.ViewCount),
		LikeCount:    uint64(post.LikeCount),
		CommentCount: uint64(post.CommentCount),
		Title:        post.Title,
		Content:      post.Content,
		ContentType:  uint32(post.ContentType),
		CreatedAt:    post.CreatedAt.Format(time.RFC3339),
		UserID:       post.UserID,
		Tags:         tags,
	}

}

func ToPostBrief(post *model2.Post) *post_grpc.PostBrief {
	return &post_grpc.PostBrief{
		ID:        post.ID,
		Title:     post.Title,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
		UserID:    post.UserID,
	}
}

// ToTopPost model 转 post_grpc.TopPost
func ToTopPost(post *model2.Post, score float64) *post_grpc.TopPost {
	return &post_grpc.TopPost{
		ID:    post.ID,
		Title: post.Title,
		Score: float32(score),
	}
}

// ToComment model 转 post_grpc.Comment
func ToComment(comment *model2.Comment) *post_grpc.Comment {
	return &post_grpc.Comment{
		ID:        comment.ID,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		ReplyID:   comment.ReplyID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}
}
