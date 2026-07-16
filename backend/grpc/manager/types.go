package manager

import (
	"context"

	agent_grpc "github.com/yzletter/go-postery/api/proto/agent/v1"
	auth_grpc "github.com/yzletter/go-postery/api/proto/auth/v1"
	code_grpc "github.com/yzletter/go-postery/api/proto/code/v1"
	inter_grpc "github.com/yzletter/go-postery/api/proto/interactive/v1"
	interview_grpc "github.com/yzletter/go-postery/api/proto/interview/v1"
	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	oss_grpc "github.com/yzletter/go-postery/api/proto/oss/v1"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	rank_grpc "github.com/yzletter/go-postery/api/proto/rank/v1"
	search_grpc "github.com/yzletter/go-postery/api/proto/search/v1"
	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	ws_gateway_grpc "github.com/yzletter/go-postery/api/proto/ws_gateway/v1"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	InterviewService   = "interview_service"
	AgentService       = "agent_service"
	AuthService        = "auth_service"
	CodeService        = "code_service"
	InteractiveService = "interactive_service"
	LotteryService     = "lottery_service"
	OSSService         = "oss_service"
	PostService        = "post_service"
	RankService        = "rank_service"
	SearchService      = "search_service"
	SessionService     = "session_service"
	UserService        = "user_service"
	WSGatewayService   = "ws_gateway_service"
)

// isEndpointFailure 判断错误是否代表当前节点不可用, 业务错误不应该降低节点健康度
func isEndpointFailure(err error) bool {
	if err == nil {
		return false
	}

	switch status.Code(err) {
	case codes.Unavailable,
		codes.DeadlineExceeded,
		codes.ResourceExhausted,
		codes.Unknown,
		codes.Internal:
		return true
	default:
		return false
	}
}

type ServiceHub interface {
	LoadEndpoints(ctx context.Context, service string)                                           // 从服务注册中心初始化所有可用连接
	AddEndpoint(ctx context.Context, service string, addr string)                                // 建立新连接
	RemoveEndpoint(ctx context.Context, service string, addr string)                             // 删除节点连接
	Take(ctx context.Context, service string) *hub.Endpoint                                      // 根据负载均衡选择一个连接
	WatchEndpointsFromServiceHub(ctx context.Context, service string)                            // Watch 一个服务
	Register(ctx context.Context, service string, endpoint string, leaseID int64) (int64, error) // 下游服务向 ServiceHub 注册 / 续约服务
	Unregister(ctx context.Context, service string, endpoint string) error                       // 下游服务向 ServiceHub 取消注册服务
	GetEndpoints(ctx context.Context, service string) []*hub.Endpoint                            // 获取所有可用节点
}

type CodeClient interface {
	Send(ctx context.Context, req *code_grpc.SendCodeRequest) (*code_grpc.SendCodeResponse, error)
	Verify(ctx context.Context, req *code_grpc.VerifyCodeRequest) (*code_grpc.VerifyCodeResponse, error)
}

type AuthClient interface {
	Login(ctx context.Context, req *auth_grpc.LoginRequest) (*auth_grpc.LoginResponse, error)
	HasPassword(ctx context.Context, req *auth_grpc.UserID) (*auth_grpc.HasPasswordResponse, error)
	SetPassword(ctx context.Context, req *auth_grpc.SetPasswordRequest) (*auth_grpc.AuthEmptyResponse, error)
	UpdatePassword(ctx context.Context, req *auth_grpc.UpdatePasswordRequest) (*auth_grpc.AuthEmptyResponse, error)
	GetAuthIdentityByUID(ctx context.Context, req *auth_grpc.UserID) (*auth_grpc.AuthIdentity, error)
	IssueTokens(ctx context.Context, req *auth_grpc.IssueTokenRequest) (*auth_grpc.DualTokens, error)
	ClearTokens(ctx context.Context, req *auth_grpc.DualTokens) (*auth_grpc.AuthEmptyResponse, error)
	VerifyAccessToken(ctx context.Context, req *auth_grpc.AccessToken) (*auth_grpc.JWTTokenClaims, error)
	GetInfoByRefreshToken(ctx context.Context, req *auth_grpc.RefreshToken) (*auth_grpc.GetInfoByRefreshTokenResponse, error)
	CheckBlackList(ctx context.Context, req *auth_grpc.CheckBlackListRequest) (*auth_grpc.CheckBlackListResponse, error)
}

type LotteryClient interface {
	GetAllGifts(ctx context.Context, req *lottery_grpc.EmptyRequest) (*lottery_grpc.Gifts, error)
	Lottery(ctx context.Context, req *lottery_grpc.UserID) (*lottery_grpc.LotteryResponse, error)
	Pay(ctx context.Context, req *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error)
	GiveUp(ctx context.Context, req *lottery_grpc.LotteryCommonRequest) (*lottery_grpc.EmptyResponse, error)
	Result(ctx context.Context, req *lottery_grpc.UserID) (*lottery_grpc.Order, error)
}

type PostClient interface {
	Create(ctx context.Context, req *post_grpc.CreatePostRequest) (*post_grpc.PostDetail, error)
	GetDetailByID(ctx context.Context, req *post_grpc.GetDetailByIDRequest) (*post_grpc.PostDetail, error)
	GetBriefByID(ctx context.Context, req *post_grpc.GetBriefByIDRequest) (*post_grpc.PostBrief, error)
	Top(ctx context.Context, req *post_grpc.PostEmptyRequest) (*post_grpc.TopResponse, error)
	GetPostByTime(ctx context.Context, req *post_grpc.GetPostByTimeRequest) (*post_grpc.PostIDs, error)
	Update(ctx context.Context, req *post_grpc.UpdateRequest) (*post_grpc.PostEmptyResponse, error)
	ListByPage(ctx context.Context, req *post_grpc.ListByPageRequest) (*post_grpc.PostDetailsResponse, error)
	ListByPageAndUid(ctx context.Context, req *post_grpc.ListByPageAndUidRequest) (*post_grpc.PostBriefsResponse, error)
	ListByPageAndTag(ctx context.Context, req *post_grpc.ListByPageAndTagRequest) (*post_grpc.PostDetailsResponse, error)
	Belong(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.BelongResponse, error)
	Delete(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error)
	ExistPost(ctx context.Context, in *post_grpc.ExistPostRequest, opts ...grpc.CallOption) (*post_grpc.ExistPostResponse, error)
	CheckPostAuth(ctx context.Context, in *post_grpc.CheckPostAuthRequest, opts ...grpc.CallOption) (*post_grpc.CheckPostAuthResponse, error)
}

type SearchClient interface {
	Search(ctx context.Context, req *search_grpc.SearchRequest) (*search_grpc.SearchResult, error)
	DeleteDoc(ctx context.Context, req *search_grpc.DocID) (*search_grpc.AffectedCount, error)
	AddDoc(ctx context.Context, req *search_grpc.Document) (*search_grpc.AffectedCount, error)
	Count(ctx context.Context, req *search_grpc.CountRequest) (*search_grpc.AffectedCount, error)
}

type AgentClient interface {
	Chat(ctx context.Context, req *agent_grpc.ChatRequest) (*agent_grpc.ChatResponse, error)
}

type InterviewClient interface {
	Chat(ctx context.Context, req *interview_grpc.ChatRequest) (*interview_grpc.ChatResponse, error)
	StartInterview(ctx context.Context, req *interview_grpc.StartInterviewRequest) (*interview_grpc.StartInterviewResponse, error)
	Answer(ctx context.Context, req *interview_grpc.AnswerRequest) (*interview_grpc.AnswerResponse, error)
	UploadQuestionsSign(ctx context.Context, req *interview_grpc.UploadQuestionsSignRequest) (*interview_grpc.UploadQuestionsSignResponse, error)
	UploadQuestionsCallback(ctx context.Context, req *interview_grpc.UploadQuestionsCallbackRequest) (*interview_grpc.UploadQuestionsCallbackResponse, error)
	UploadQuestions(ctx context.Context, req *interview_grpc.UploadQuestionsRequest) (*interview_grpc.UploadQuestionsResponse, error)
	QuitInterview(ctx context.Context, req *interview_grpc.QuitInterviewRequest) (*interview_grpc.QuitInterviewResponse, error)
	Evaluation(ctx context.Context, req *interview_grpc.EvaluationRequest) (*interview_grpc.EvaluationResponse, error)
}

type UserClient interface {
	GetProfile(ctx context.Context, req *user_grpc.GetProfileByIdRequest) (*user_grpc.Profile, error)
	UpdateProfile(ctx context.Context, req *user_grpc.UpdateProfileRequest) (*user_grpc.UpdateProfileResponse, error)
	Top(ctx context.Context, req *user_grpc.TopRequest) (*user_grpc.TopResponse, error)
	GetIDAfterTime(ctx context.Context, req *user_grpc.GetIDAfterTimeRequest) (*user_grpc.UserIDs, error)
	ListFollowersByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error)
	ListFolloweesByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error)
	UploadAvatarSign(ctx context.Context, req *user_grpc.UploadAvatarSignRequest) (*user_grpc.UploadAvatarSignResponse, error)
	UploadAvatarCallback(ctx context.Context, req *user_grpc.UploadAvatarCallbackRequest) (*user_grpc.UploadAvatarCallbackResponse, error)
	GetAvatarURL(ctx context.Context, req *user_grpc.GetAvatarURLRequest) (*user_grpc.GetAvatarURLResponse, error)
}

type OSSClient interface {
	SignUpload(ctx context.Context, req *oss_grpc.SignUploadRequest) (*oss_grpc.SignUploadResponse, error)
	UploadCallback(ctx context.Context, req *oss_grpc.UploadCallbackRequest) (*oss_grpc.UploadCallbackResponse, error)
	GetObjectURL(ctx context.Context, req *oss_grpc.GetObjectURLRequest) (*oss_grpc.GetObjectURLResponse, error)
}

type WSGatewayClient interface {
	Push(ctx context.Context, req *ws_gateway_grpc.PushRequest) (*ws_gateway_grpc.PushResponse, error)
}

type SessionClient interface {
	NewConnection(ctx context.Context, req *session_grpc.UserID) (*session_grpc.SessionEmptyResponse, error)
	Chat(ctx context.Context, req *session_grpc.ChatRequest) (*session_grpc.SessionEmptyResponse, error)
	ListByUID(ctx context.Context, req *session_grpc.UserID) (*session_grpc.Sessions, error)
	GetSession(ctx context.Context, req *session_grpc.BothUserID) (*session_grpc.Session, error)
	GetHistoryMessagesByPage(ctx context.Context, req *session_grpc.GetHistoryMessagesByPageRequest) (*session_grpc.GetHistoryMessagesByPageResponse, error)
	Delete(ctx context.Context, req *session_grpc.DeleteRequest) (*session_grpc.SessionEmptyResponse, error)
	UpdateUnread(ctx context.Context, req *session_grpc.UpdateUnreadRequest) (*session_grpc.SessionEmptyResponse, error)
	ClearUnread(ctx context.Context, req *session_grpc.ClearUnreadRequest) (*session_grpc.SessionEmptyResponse, error)
	CreateMessage(ctx context.Context, req *session_grpc.Message) (*session_grpc.Message, error)
}

type InteractiveClient interface {
	GetPostInteractive(context.Context, *inter_grpc.PostIDRequest) (*inter_grpc.PostInteractive, error)
	GetUserInteractive(context.Context, *inter_grpc.UserIDRequest) (*inter_grpc.UserInteractive, error)
	Like(context.Context, *inter_grpc.LikeRequest) (*inter_grpc.InteractiveEmptyResponse, error)
	Unlike(context.Context, *inter_grpc.LikeRequest) (*inter_grpc.InteractiveEmptyResponse, error)
	CheckLike(context.Context, *inter_grpc.LikeRequest) (*inter_grpc.CheckLikeResponse, error)
	Follow(context.Context, *inter_grpc.FollowRequest) (*inter_grpc.InteractiveEmptyResponse, error)
	Unfollow(context.Context, *inter_grpc.FollowRequest) (*inter_grpc.InteractiveEmptyResponse, error)
	IfFollow(context.Context, *inter_grpc.FollowRequest) (*inter_grpc.IfFollowResponse, error)
	Comment(context.Context, *inter_grpc.CreateCommentRequest) (*inter_grpc.InteractiveComment, error)
	DelComment(context.Context, *inter_grpc.DeleteCommentRequest) (*inter_grpc.InteractiveEmptyResponse, error)
	ListCommentByPage(context.Context, *inter_grpc.ListCommentByPageRequest) (*inter_grpc.CommentsResponse, error)
	ListRepliesByPage(context.Context, *inter_grpc.ListReplyByPageRequest) (*inter_grpc.CommentsResponse, error)
	CheckCommentDelAuth(context.Context, *inter_grpc.CommentIDUserIDRequest) (*inter_grpc.CheckCommentDelAuthResponse, error)
	GetFollowers(context.Context, *inter_grpc.ListFollowRequest) (*inter_grpc.ListFollowResponse, error)
	GetFollowees(context.Context, *inter_grpc.ListFollowRequest) (*inter_grpc.ListFollowResponse, error)
}

type RankClient interface {
	RankUser(context.Context, *rank_grpc.RankIDRequest) (*rank_grpc.RankEmptyResponse, error)
	RankPost(context.Context, *rank_grpc.RankIDRequest) (*rank_grpc.RankEmptyResponse, error)
	RankTopKUser(context.Context, *rank_grpc.RankEmptyRequest) (*rank_grpc.RankEmptyResponse, error)
	RankTopKPost(context.Context, *rank_grpc.RankEmptyRequest) (*rank_grpc.RankEmptyResponse, error)
	TopKUser(context.Context, *rank_grpc.RankEmptyRequest) (*rank_grpc.TopKUserResponse, error)
	TopKPost(context.Context, *rank_grpc.RankEmptyRequest) (*rank_grpc.TopKPostResponse, error)
}
