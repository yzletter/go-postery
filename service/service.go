package service

import (
	"context"
	"net/http"

	commentdto "github.com/yzletter/go-postery/dto/comment"
	giftdto "github.com/yzletter/go-postery/dto/gift"
	messagedto "github.com/yzletter/go-postery/dto/message"
	orderdto "github.com/yzletter/go-postery/dto/order"
	postdto "github.com/yzletter/go-postery/dto/post"
	sessiondto "github.com/yzletter/go-postery/dto/session"
	userdto "github.com/yzletter/go-postery/dto/user"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/service/ports"
)

// 定义 Service 层所有接口

type AuthService interface {
	LoginByPassword(ctx context.Context, identifier, password string) (userdto.BriefDTO, error) // 手机号码/邮箱 + 密码登录
	LoginByPhone(ctx context.Context, phone, code string) (userdto.BriefDTO, error)             // 手机号码 + 验证码进行登录, 未注册的手机号码自动进行注册
	HasPassword(ctx context.Context, uid int64) (bool, error)                                   // 查询密码状态
	SetPassword(ctx context.Context, uid int64, code, newPass string) error                     // 初始化密码
	UpdatePassword(ctx context.Context, uid int64, oldPass, newPass string) error               // 修改密码
	GetAuthIdentityByUID(ctx context.Context, uid int64) (string, string, error)                // 获取用户身份认证
	IssueTokens(ctx context.Context, id int64, role int, agent string) (string, string, error)  // 签发双 Token
	ClearTokens(ctx context.Context, accessToken, refreshToken string) error                    // 清除双 Token
	VerifyAccessToken(tokenString string) (*ports.JWTTokenClaims, error)                        // 校验 AccessToken
	GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error) // 根据 RefreshToken 获取用户信息, 用于重新签发双 Token
	CheckBlackList(ctx context.Context, ssid string) (bool, error)                              // 根据 SSID 检查黑名单, 检查用户是否被拉黑
}

type CodeService interface {
	SendCode(ctx context.Context, biz model.CodeBiz, field string) error                       // 发送验证码
	CheckCode(ctx context.Context, biz model.CodeBiz, field string, code string) (bool, error) // 校验验证码
}

type UserService interface {
	GetProfileById(ctx context.Context, id int64) (userdto.DetailDTO, error)
	UpdateProfile(ctx context.Context, id int64, req userdto.ModifyProfileRequest) error
	Top(ctx context.Context) ([]userdto.TopDTO, error)
}

type PostService interface {
	Create(ctx context.Context, uid int64, title string, content string, contentType int) (postdto.DetailDTO, error)
	GetDetailById(ctx context.Context, id int64, addViewCnt bool) (postdto.DetailDTO, error)
	GetBriefById(ctx context.Context, id int64) (postdto.BriefDTO, error)
	Belong(ctx context.Context, pid, uid int64) bool
	Delete(ctx context.Context, pid, uid int64) error
	Update(ctx context.Context, pid int64, uid int64, title, content string, tags []string) error
	ListByPage(ctx context.Context, pageNo, pageSize int) (int, []postdto.DetailDTO, error)
	ListByPageAndUid(ctx context.Context, uid int64, pageNo, pageSize int) (int, []postdto.BriefDTO, error)
	ListByPageAndTag(ctx context.Context, name string, pageNo, pageSize int) (int, []postdto.DetailDTO, error)
	Like(ctx context.Context, pid, uid int64) error
	Unlike(ctx context.Context, pid, uid int64) error
	IfLike(ctx context.Context, pid, uid int64) (bool, error)
	Top(ctx context.Context) ([]postdto.TopDTO, error)
}

type CommentService interface {
	Create(ctx context.Context, pid int64, uid int64, parentId int64, replyId int64, content string) (commentdto.DTO, error)
	Delete(ctx context.Context, uid, cid int64) error
	List(ctx context.Context, pid int64, pageNo, pageSize int) (int, []commentdto.DTO, error)
	ListReplies(ctx context.Context, ids int64, pageNo, pageSize int) (int, []commentdto.DTO, error)
	CheckAuth(ctx context.Context, cid, uid int64) bool
}

type TagService interface {
	Create(ctx context.Context, name string) (int64, error)
	Bind(ctx context.Context, pid int64, tags []string) error
	FindTagsByPostID(ctx context.Context, pid int64) ([]string, error)
}

type FollowService interface {
	StartInitUserScoreConsumer(ctx context.Context)
	Follow(ctx context.Context, ferId, feeId int64) error
	UnFollow(ctx context.Context, ferId, feeId int64) error
	IfFollow(ctx context.Context, ferId, feeId int64) (model.FollowType, error)
	ListFollowersByPage(ctx context.Context, uid int64, pageNo, pageSize int) (int, []userdto.BriefDTO, error)
	ListFolloweesByPage(ctx context.Context, uid int64, pageNo, pageSize int) (int, []userdto.BriefDTO, error)
}

type SessionService interface {
	ListByUid(ctx context.Context, uid int64) ([]sessiondto.DTO, error)
	GetSession(ctx context.Context, uid, targetID int64) (sessiondto.DTO, error)
	Register(ctx context.Context, uid int64) error
	GetHistoryMessagesByPage(ctx context.Context, uid int64, targetID int64, pageNo, pageSize int) (int, []messagedto.DTO, error)
	Delete(ctx context.Context, uid, sid int64) error
	StartSessionRegisterConsumer(ctx context.Context)
}

type WebsocketService interface {
	Connect(ctx context.Context, w http.ResponseWriter, r *http.Request, uid int64) error
}

type LotteryService interface {
	GetAllGifts(ctx context.Context) ([]giftdto.DTO, error)
	Lottery(ctx context.Context, uid int64) (giftdto.DTO, error)
	Pay(ctx context.Context, uid, gid int64) error
	GiveUp(ctx context.Context, uid, gid int64) error
	Result(ctx context.Context, uid int64) (orderdto.DTO, error)
	StartLotteryOrderConsumer(ctx context.Context)
	InitCacheInventory(ctx context.Context)
}

type AgentService interface {
}
