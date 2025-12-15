package repository

import "github.com/yzletter/go-postery/model"

type UserRepository interface {
	Create(user *model.User) (*model.User, error)
	Delete(id int64) error
	GetPasswordHash(id int64) (string, error)
	GetStatus(id int64) (int, error)
	GetByID(id int64) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	UpdatePasswordHash(id int64, newHash string) error
	UpdateProfile(id int64, updates map[string]any) error
}

type PostRepository interface {
	Create(uid int, title, content string) (model.Post, error)
	Delete(pid int) error
	Update(pid int, title, content string) error
	GetByID(pid int) (bool, model.Post)
	GetByPage(pageNo, pageSize int) (int, []model.Post)
	GetByPageAndTag(tid, pageNo, pageSize int) (int, []model.Post)
	GetByUid(uid int) []model.Post
	ChangeViewCnt(pid int, delta int)
	ChangeLikeCnt(pid int, delta int)
	ChangeCommentCnt(pid int, delta int)
}
type CommentRepository interface {
}
type TagRepository interface {
}
type LikeRepository interface {
}
type FollowRepository interface {
}
