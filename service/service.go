package service

import (
	"context"

	userdto "github.com/yzletter/go-postery/dto/user"
)

// 定义 Service 层所有接口

type UserService interface {
	Register(ctx context.Context, username, email, password string) (userdto.BriefDTO, error)
	GetBriefById(ctx context.Context, id int64) (userdto.BriefDTO, error)
	GetDetailById(ctx context.Context, id int64) (userdto.DetailDTO, error)
	GetBriefByName(ctx context.Context, username string) (userdto.BriefDTO, error)
	UpdatePassword(ctx context.Context, id int64, oldPass, newPass string) error
	UpdateProfile(ctx context.Context, id int64, req userdto.ModifyProfileRequest) error
	Login(ctx context.Context, username, pass string) (userdto.BriefDTO, error)
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

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (userdto.BriefDTO, error)
	Login(ctx context.Context, username, pass string) (userdto.BriefDTO, error)
	IssueTokens(ctx context.Context, id int64, role int) (string, string, error)
	Logout(ctx context.Context, accessToken string) error
}
