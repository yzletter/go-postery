package service

import (
	"context"

	"github.com/yzletter/go-postery/dto/comment"
	"github.com/yzletter/go-postery/dto/post"
	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/model"
)

// 定义 Service 层所有接口

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (userdto.BriefDTO, error)
	Login(ctx context.Context, username, pass string) (userdto.BriefDTO, error)
	ClearTokens(ctx context.Context, accessToken, refreshToken string) error
	IssueTokens(ctx context.Context, id int64, role int, agent string) (string, string, error)
	VerifyAccessToken(tokenString string) (*JWTTokenClaims, error)
}

type UserService interface {
	GetBriefById(ctx context.Context, id int64) (userdto.BriefDTO, error)
	GetDetailById(ctx context.Context, id int64) (userdto.DetailDTO, error)
	GetBriefByName(ctx context.Context, username string) (userdto.BriefDTO, error)
	UpdatePassword(ctx context.Context, id int64, oldPass, newPass string) error
	UpdateProfile(ctx context.Context, id int64, req userdto.ModifyProfileRequest) error
}

type PostService interface {
	Create(ctx context.Context, uid int64, title, content string) (post.DetailDTO, error)
	GetDetailById(ctx context.Context, id int64, addViewCnt bool) (post.DetailDTO, error)
	GetBriefById(ctx context.Context, id int64) (post.BriefDTO, error)
	Belong(ctx context.Context, pid, uid int64) bool
	Delete(ctx context.Context, pid, uid int64) error
	Update(ctx context.Context, pid int64, uid int64, title, content string, tags []string) error
	ListByPage(ctx context.Context, pageNo, pageSize int) (int, []post.DetailDTO)
	ListByUid(ctx context.Context, uid int64, pageNo, pageSize int) (int, []post.BriefDTO)
	ListByPageAndTag(ctx context.Context, name string, pageNo, pageSize int) (int, []post.DetailDTO)
}

type CommentService interface {
	Create(ctx context.Context, pid int64, uid int64, parentId int64, replyId int64, content string) (comment.DTO, error)
	Delete(ctx context.Context, uid, cid int64) error
	List(ctx context.Context, pid int64, pageNo, pageSize int) (int, []comment.DTO)
	CheckAuth(ctx context.Context, cid, uid int64) bool
}

type TagService interface {
	Create(ctx context.Context, name string) (int64, error)
	Bind(ctx context.Context, pid int64, tags []string) error
	FindTagsByPostID(ctx context.Context, pid int64) ([]string, error)
}

type FollowService interface {
	Follow(ctx context.Context, ferId, feeId int64) error
	UnFollow(ctx context.Context, ferId, feeId int64) error
	IfFollow(ctx context.Context, ferId, feeId int64) (model.FollowType, error)
	GetFollowersByPage(ctx context.Context, uid int64, pageNo, pageSize int) (int, []userdto.BriefDTO, error)
	GetFolloweesByPage(ctx context.Context, uid int64, pageNo, pageSize int) (int, []userdto.BriefDTO, error)
}
