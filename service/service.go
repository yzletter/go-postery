package service

import (
	"context"

	"github.com/yzletter/go-postery/dto/request"
	dto "github.com/yzletter/go-postery/dto/response"
)

// 定义 Service 层所有接口

type UserService interface {
	Register(ctx context.Context, username, email, password string) (dto.UserBriefDTO, error)
	GetBriefById(ctx context.Context, id int64) (dto.UserBriefDTO, error)
	GetDetailById(ctx context.Context, id int64) (dto.UserDetailDTO, error)
	GetBriefByName(ctx context.Context, username string) (dto.UserBriefDTO, error)
	UpdatePassword(ctx context.Context, id int64, oldPass, newPass string) error
	UpdateProfile(ctx context.Context, id int64, req request.ModifyProfileRequest) error
	Login(ctx context.Context, username, pass string) (dto.UserBriefDTO, error)
}

type PostService interface {
}

type CommentService interface {
}

type LikeService interface {
}

type FollowService interface {
}

type TagService interface {
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, plainPassword string) error
}
