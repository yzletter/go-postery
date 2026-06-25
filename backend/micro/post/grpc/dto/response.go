package dto

import (
	"time"

	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"github.com/yzletter/go-postery/backend/micro/post/domain"
)

// ToPostDetail domain.Post 转 post_grpc.PostDetail
func ToPostDetail(post domain.Post) *post_grpc.PostDetail {
	return &post_grpc.PostDetail{
		ID:           post.ID,
		ViewCount:    uint64(post.ViewCount),
		LikeCount:    uint64(post.LikeCount),
		CommentCount: uint64(post.CommentCount),
		Title:        post.Title,
		Content:      post.Content,
		ContentType:  uint32(post.ContentType),
		CreatedAt:    post.CreatedAt.Format(time.RFC3339),
		UserID:       post.User.UserID,
		Tags:         post.Tags,
	}

}

// ToPostBrief domain.PostBrief 转 post_grpc.PostBrief
func ToPostBrief(post domain.PostBrief) *post_grpc.PostBrief {
	return &post_grpc.PostBrief{
		ID:        post.ID,
		Title:     post.Title,
		Abstract:  post.Abstract,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
		UserID:    post.User.UserID,
	}
}

// ToTopPost domain.PostBrief 转 post_grpc.TopPost
func ToTopPost(post domain.PostBrief, score float64) *post_grpc.TopPost {
	return &post_grpc.TopPost{
		ID:    post.ID,
		Title: post.Title,
		Score: float32(score),
	}
}
