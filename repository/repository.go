package repository

import (
	"context"

	"github.com/yzletter/go-postery/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int64) error
	GetPasswordHash(ctx context.Context, id int64) (string, error)
	GetStatus(ctx context.Context, id int64) (int, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	UpdatePasswordHash(ctx context.Context, id int64, newHash string) error
	UpdateProfile(ctx context.Context, id int64, updates map[string]any) error
}

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) (*model.Post, error)
	Delete(ctx context.Context, id int64) error
	UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error
	Update(ctx context.Context, id int64, updates map[string]any) error
	GetByID(ctx context.Context, id int64) (*model.Post, error)
	GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error)
	GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error)
	GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	GetByID(ctx context.Context, id int64) (*model.Comment, error)
	Delete(ctx context.Context, id int64) (int, error)
	GetByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Comment, error)
	GetRepliesByParentIDs(ctx context.Context, ids []int64) ([]*model.Comment, error)
}

type LikeRepository interface {
	Like(ctx context.Context, like *model.Like) error
	UnLike(ctx context.Context, uid, pid int64) error
	HasLiked(ctx context.Context, uid, pid int64) (bool, error)
}

type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	GetBySlug(ctx context.Context, slug string) (*model.Tag, error)
	GetByName(ctx context.Context, name string) (*model.Tag, error)
	Bind(ctx context.Context, postTag *model.PostTag) error
	DeleteBind(ctx context.Context, pid, tid int64) error
	FindTagsByPostID(ctx context.Context, pid int64) ([]string, error)
}

type FollowRepository interface {
	Create(ctx context.Context, follow *model.Follow) error
	Delete(ctx context.Context, ferID, feeID int64) error
	Exists(ctx context.Context, ferID, feeID int64) (int, error)
	GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)
	GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)
}
