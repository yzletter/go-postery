package repository

import (
	"context"

	"github.com/yzletter/go-postery/dto/session"
	"github.com/yzletter/go-postery/model"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, authAggregate *model.AuthAggregate) error                          // 创建用户（包括用户最小项、用户登录认证、用户密码、用户资料、注册扩展功能）
	GetAuthIdentity(ctx context.Context, authType int, identifier string) (*model.AuthIdentity, error) // 根据登录方式和凭证获取登录认证
	GetPasswordHash(ctx context.Context, uid int64) (string, error)                                    // 根据 UID 获取用户密码
	DelRefreshToken(ctx context.Context, refreshToken string) error                                    // 缓存中删除 RefreshToken
	SetInfo(ctx context.Context, refreshToken string, mp map[string]any) error                         // 根据 RefreshToken 在缓存中存储用户信息
	GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error)        // 根据 RefreshToken 从缓存中读取用户信息
	SetBlackList(ctx context.Context, ssid string) error                                               // 拉黑 SSID
	CheckBlackList(ctx context.Context, ssid string) (bool, error)                                     // 查看 SSID 是否被拉黑
}

type UserRepository interface {
	GetProfileByID(ctx context.Context, uid int64) (*model.UserProfile, error) // 根据 ID 查找用户资料
	UpdateProfile(ctx context.Context, id int64, updates map[string]any) error // 根据 ID 修改用户资料的多个字段
	Top(ctx context.Context) ([]*model.UserProfile, []float64, error)          // 返回热门推荐用户
	ChangeScore(ctx context.Context, uid int64, delta int)                     // 修改用户分数

	//Create(ctx context.Context, user *model.User) error
	//Delete(ctx context.Context, id int64) error
	//GetPasswordHash(ctx context.Context, id int64) (string, error)
	//GetStatus(ctx context.Context, id int64) (int, error)
	//GetProfileByID(ctx context.Context, id int64) (*model.User, error)
	//GetByUsername(ctx context.Context, username string) (*model.User, error)
	//GetByEmail(ctx context.Context, email string) (*model.User, error)
	//GetByPhone(ctx context.Context, phone string) (*model.User, error)
	//UpdatePasswordHash(ctx context.Context, id int64, newHash string) error
}

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id int64) error
	UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error
	Update(ctx context.Context, id int64, updates map[string]any) error
	GetByID(ctx context.Context, id int64) (*model.Post, error)
	GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error)
	GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error)
	GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error)
	ChangeScore(ctx context.Context, pid int64, delta int)
	Top(ctx context.Context) ([]*model.Post, []float64, error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByID(ctx context.Context, id int64) (*model.Comment, error)
	Delete(ctx context.Context, id int64) (int, error)
	GetByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Comment, error)
	GetRepliesByParentID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Comment, error)
}

type LikeRepository interface {
	Like(ctx context.Context, like *model.Like) error
	UnLike(ctx context.Context, uid, pid int64) error
	HasLiked(ctx context.Context, uid, pid int64) (bool, error)
}

type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	GetBySlug(ctx context.Context, slug string) (*model.Tag, error)
	GetByName(ctx context.Context, name string) (*model.Tag, error)
	Bind(ctx context.Context, postTag *model.PostTag) error
	DeleteBind(ctx context.Context, pid, tid int64) error
	FindTagsByPostID(ctx context.Context, pid int64) ([]string, error)
}

type FollowRepository interface {
	Create(ctx context.Context, follow *model.Follow) error
	Delete(ctx context.Context, ferID, feeID int64) error
	Exists(ctx context.Context, ferID, feeID int64) (model.FollowType, error)
	GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)
	GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	ListByUid(ctx context.Context, uid int64) ([]*model.Session, error)
	GetByUidAndTargetID(ctx context.Context, uid, targetID int64) (*model.Session, error)
	GetByID(ctx context.Context, uid, sid int64) (*model.Session, error)
	Delete(ctx context.Context, uid, sid int64) error
	UpdateUnread(ctx context.Context, uid int64, sid int64, updates session.UpdateUnreadRequest) error
	ClearUnread(ctx context.Context, uid int64, sid int64) error
}

type MessageRepository interface {
	Create(ctx context.Context, message *model.Message) error
	GetByIDAndTargetID(ctx context.Context, id, targetID int64) ([]*model.Message, error)
	GetByPage(ctx context.Context, id int64, targetID int64, pageNo, pageSize int) (int, []*model.Message, error)
}

type OrderRepository interface {
	CreateTempOrder(ctx context.Context, uid, gid int64) error
	DeleteTempOrder(ctx context.Context, uid int64) error
	GetTempOrder(ctx context.Context, uid int64) (int64, error)
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrder(ctx context.Context, uid int64) (*model.Order, error)
}

type GiftRepository interface {
	GetAllGifts(ctx context.Context) ([]*model.Gift, error)
	GetCacheInventory(ctx context.Context) ([]*model.Gift, error)
	GetByID(ctx context.Context, gid int64) (*model.Gift, error)
	ReduceCacheInventory(ctx context.Context, gid int64) error
	IncreaseCacheInventory(ctx context.Context, gid int64) error
	InitCacheInventory(ctx context.Context)
}

type AgentRepository interface {
}

type CodeRepository interface {
	Allow(ctx context.Context, biz model.CodeBiz, field string, code string) error             // Allow 判断是否允许发送 Code
	CheckCode(ctx context.Context, biz model.CodeBiz, field string, code string) (bool, error) // CheckCode 校验 Code
}
