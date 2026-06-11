package dao

import (
	"context"

	model2 "github.com/yzletter/go-postery/backend/micro/post/model"
)

type PostDAO interface {
	Create(ctx context.Context, post *model2.Post, events []*model2.Event) error
	Delete(ctx context.Context, id int64) error
	UpdateCount(ctx context.Context, id int64, field model2.PostCntField, delta int) error
	Update(ctx context.Context, id int64, updates map[string]any) error
	GetByID(ctx context.Context, id int64) (*model2.Post, error)
	GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model2.Post, error)
	GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model2.Post, error)
	GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model2.Post, error)
}

type LikeDAO interface {
	Create(ctx context.Context, like *model2.Like) error
	Delete(ctx context.Context, uid, pid int64) error
	Exists(ctx context.Context, uid, pid int64) (bool, error)
}

type TagDAO interface {
	Create(ctx context.Context, tag *model2.Tag) error
	GetBySlug(ctx context.Context, slug string) (*model2.Tag, error)
	GetByName(ctx context.Context, name string) (*model2.Tag, error)
	Bind(ctx context.Context, postTag *model2.PostTag) error
	DeleteBind(ctx context.Context, pid, tid int64) error
	FindTagsByPostID(ctx context.Context, pid int64) ([]string, error)
}

type CommentDAO interface {
	Create(ctx context.Context, comment *model2.Comment) error
	Delete(ctx context.Context, id int64) (int, error)
	GetByID(ctx context.Context, id int64) (*model2.Comment, error)
	GetByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model2.Comment, error)
	GetRepliesByParentID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model2.Comment, error)
}
