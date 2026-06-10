package client

import (
	"context"

	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	hub2 "github.com/yzletter/go-postery/microservice-backend/search/grpc/hub"
	search_model "github.com/yzletter/go-postery/microservice-backend/search/model"
)

type ServiceHub interface {
	LoadEndpoints(ctx context.Context, service string)
	WatchEndpointsFromServiceHub(ctx context.Context, service string)
	Take(ctx context.Context, service string) *hub2.Endpoint
}

const (
	CodeServiceName    = "code_service"
	AuthServiceName    = "auth_service"
	LotteryServiceName = "lottery_service"
	PostServiceName    = "post_service"
	SearchServiceName  = "search_service"
	AgentServiceName   = "agent_service"
	UserServiceName    = "user_service"
	SessionServiceName = "session_service"
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
	Lottery(ctx context.Context, req *lottery_grpc.UserID) (*lottery_grpc.LotteryResponse, error)
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

type SearchClient interface {
	Search(ctx context.Context, req *search_grpc.SearchRequest) (*search_grpc.SearchResult, error)
	DeleteDoc(ctx context.Context, req *search_grpc.DocID) (*search_grpc.AffectedCount, error)
	AddDoc(ctx context.Context, req *search_model.Document) (*search_grpc.AffectedCount, error)
	Count(ctx context.Context, req *search_grpc.CountRequest) (*search_grpc.AffectedCount, error)
	Close()
}

type AgentClient interface {
	Chat(ctx context.Context, req *agent_grpc.ChatRequest) (*agent_grpc.ChatResponse, error)
	Close()
}

type UserClient interface {
	GetProfileById(ctx context.Context, req *user_grpc.GetProfileByIdRequest) (*user_grpc.UserDetail, error)
	UpdateProfile(ctx context.Context, req *user_grpc.UpdateProfileRequest) (*user_grpc.UpdateProfileResponse, error)
	Top(ctx context.Context, req *user_grpc.TopRequest) (*user_grpc.TopResponse, error)
	Follow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.FollowEmptyResponse, error)
	UnFollow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.FollowEmptyResponse, error)
	IfFollow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.IfFollowResponse, error)
	ListFollowersByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error)
	ListFolloweesByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error)
	UploadAvatarSign(ctx context.Context, req *user_grpc.UploadAvatarSignRequest) (*user_grpc.UploadAvatarSignResponse, error)
	UploadAvatarCallback(ctx context.Context, req *user_grpc.UploadAvatarCallbackRequest) (*user_grpc.UploadAvatarCallbackResponse, error)
	GetAvatarURL(ctx context.Context, req *user_grpc.GetAvatarURLRequest) (*user_grpc.GetAvatarURLResponse, error)
	Close()
}

type SessionClient interface {
	ListByUID(ctx context.Context, req *session_grpc.UserID) (*session_grpc.Sessions, error)
	GetSession(ctx context.Context, req *session_grpc.BothUserID) (*session_grpc.Session, error)
	GetHistoryMessagesByPage(ctx context.Context, req *session_grpc.GetHistoryMessagesByPageRequest) (*session_grpc.GetHistoryMessagesByPageResponse, error)
	Delete(ctx context.Context, req *session_grpc.DeleteRequest) (*session_grpc.SessionEmptyResponse, error)
	UpdateUnread(ctx context.Context, req *session_grpc.UpdateUnreadRequest) (*session_grpc.SessionEmptyResponse, error)
	ClearUnread(ctx context.Context, req *session_grpc.ClearUnreadRequest) (*session_grpc.SessionEmptyResponse, error)
	CreateMessage(ctx context.Context, req *session_grpc.Message) (*session_grpc.Message, error)
	Close()
}
