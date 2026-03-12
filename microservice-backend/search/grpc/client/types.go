package client

import (
	"context"

	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
)

const (
	CodeClientAddr    = "172.16.150.246:9001"
	AuthClientAddr    = "172.16.150.246:9002"
	LotteryClientAddr = "172.16.150.246:9003"
	PostClientAddr    = "172.16.150.246:9004"
	SearchClientAddr  = "172.16.150.246:9005"
	AgentClientAddr   = "172.16.150.246:9006"
	UserClientAddr    = "172.16.150.246:9007"
	SessionClientAddr = "172.16.150.246:9008"
)

type CodeClient interface {
	Send(ctx context.Context, req *code_grpc.SendCodeRequest) (*code_grpc.SendCodeResponse, error)
	Verify(ctx context.Context, req *code_grpc.CheckCodeRequest) (*code_grpc.CheckCodeResponse, error)
	Close()
}

type AuthClient interface {
	LoginByPassword(ctx context.Context, req *auth_grpc.LoginByPasswordRequest) (*auth_grpc.UserID, error)
	LoginByPhone(ctx context.Context, req *auth_grpc.LoginByPhoneRequest) (*auth_grpc.UserID, error)
	HasPassword(ctx context.Context, req *auth_grpc.UserID) (*auth_grpc.HasPasswordResponse, error)
	SetPassword(ctx context.Context, req *auth_grpc.SetPasswordRequest) (*auth_grpc.AuthEmptyResponse, error)
	UpdatePassword(ctx context.Context, req *auth_grpc.UpdatePasswordRequest) (*auth_grpc.AuthEmptyResponse, error)
	GetAuthIdentityByUID(ctx context.Context, req *auth_grpc.UserID) (*auth_grpc.AuthIdentity, error)
	IssueTokens(ctx context.Context, req *auth_grpc.IssueTokenRequest) (*auth_grpc.DualTokens, error)
	ClearTokens(ctx context.Context, req *auth_grpc.DualTokens) (*auth_grpc.AuthEmptyResponse, error)
	VerifyAccessToken(ctx context.Context, req *auth_grpc.AccessToken) (*auth_grpc.JWTTokenClaims, error)
	GetInfoByRefreshToken(ctx context.Context, req *auth_grpc.RefreshToken) (*auth_grpc.GetInfoByRefreshTokenResponse, error)
	CheckBlackList(ctx context.Context, req *auth_grpc.CheckBlackListRequest) (*auth_grpc.CheckBlackListResponse, error)
	Close()
}

type LotteryClient interface {
	GetAllGifts(ctx context.Context, req *lottery_grpc.EmptyRequest) (*lottery_grpc.Gifts, error)
	Lottery(ctx context.Context, req *lottery_grpc.UserID) (*lottery_grpc.Gift, error)
	Pay(ctx context.Context, req *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error)
	GiveUp(ctx context.Context, req *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error)
	Result(ctx context.Context, req *lottery_grpc.UserID) (*lottery_grpc.Order, error)
	Close()
}

type PostClient interface {
	Create(ctx context.Context, req *post_grpc.CreatePostRequest) (*post_grpc.PostDetail, error)
	GetDetailByID(ctx context.Context, req *post_grpc.GetDetailByIDRequest) (*post_grpc.PostDetail, error)
	GetBriefByID(ctx context.Context, req *post_grpc.GetBriefByIDRequest) (*post_grpc.PostBrief, error)
	Top(ctx context.Context, req *post_grpc.PostEmptyRequest) (*post_grpc.TopResponse, error)
	Update(ctx context.Context, req *post_grpc.UpdateRequest) (*post_grpc.PostEmptyResponse, error)
	ListByPage(ctx context.Context, req *post_grpc.ListByPageRequest) (*post_grpc.PostDetailsResponse, error)
	ListByPageAndUid(ctx context.Context, req *post_grpc.ListByPageAndUidRequest) (*post_grpc.PostBriefsResponse, error)
	ListByPageAndTag(ctx context.Context, req *post_grpc.ListByPageAndTagRequest) (*post_grpc.PostDetailsResponse, error)
	Belong(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.BelongResponse, error)
	Delete(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error)
	Like(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error)
	Unlike(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error)
	IfLike(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.IfLikeResponse, error)
	CreateComment(ctx context.Context, req *post_grpc.CreateCommentRequest) (*post_grpc.Comment, error)
	DeleteComment(ctx context.Context, req *post_grpc.DeleteCommentRequest) (*post_grpc.PostEmptyResponse, error)
	ListCommentByPage(ctx context.Context, req *post_grpc.ListCommentByPageRequest) (*post_grpc.CommentsResponse, error)
	ListRepliesByPage(ctx context.Context, req *post_grpc.ListReplyByPageRequest) (*post_grpc.CommentsResponse, error)
	CheckCommentDeleteAuth(ctx context.Context, req *post_grpc.CommentBelongRequest) (*post_grpc.BelongResponse, error)
	Close()
}
