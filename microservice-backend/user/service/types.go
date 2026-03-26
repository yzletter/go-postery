package service

import (
	"context"

	"github.com/yzletter/go-postery/microservice-backend/user/model"
)

type UserService interface {
	UploadAvatarSign(ctx context.Context, uid int64) (string, error)
	UploadAvatarCallback(ctx context.Context, uid int64, objectName string) error
	GetAvatarURL(ctx context.Context, objectName string) (string, error)
	GetProfileByID(ctx context.Context, id int64) (*model.UserProfile, error) // 根据用户 ID 获取用户资料
	UpdateProfile(ctx context.Context, profile *model.UserProfile) error      // 更新用户资料
	Top(ctx context.Context) ([]*model.UserProfile, []float64, error)         // 返回推荐用户
	Follow(ctx context.Context, followerID int64, followeeID int64) error
	UnFollow(ctx context.Context, followerID int64, followeeID int64) error
	IfFollow(ctx context.Context, followerID int64, followeeID int64) (int, error)
	ListFollowersByPage(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []*model.UserProfile, error) // 按页查找粉丝
	ListFolloweesByPage(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []*model.UserProfile, error) // 按页查找关注的人
	StartInitUserScoreConsumer(ctx context.Context)
}
