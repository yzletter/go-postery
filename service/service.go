package service

import (
	"context"

	dto "github.com/yzletter/go-postery/dto/user"
)

// 定义 Service 层所有接口

type UserService interface {
	Register(ctx context.Context, username, email, password string) (dto.BriefDTO, error)
	GetBriefById(ctx context.Context, id int64) (dto.BriefDTO, error)
	GetDetailById(ctx context.Context, id int64) (dto.DetailDTO, error)
	GetBriefByName(ctx context.Context, username string) (dto.BriefDTO, error)
	UpdatePassword(ctx context.Context, id int64, oldPass, newPass string) error
	UpdateProfile(ctx context.Context, id int64, req dto.ModifyProfileRequest) error
	Login(ctx context.Context, username, pass string) (dto.BriefDTO, error)
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
