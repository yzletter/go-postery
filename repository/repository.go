package repository

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
	"github.com/yzletter/go-postery/dto/session"
	"github.com/yzletter/go-postery/model"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, authAggregate *model.AuthAggregate) error                            // 创建用户（包括用户最小项、用户登录认证、用户密码、用户资料、注册扩展功能）
	GetAuthIdentity(ctx context.Context, authType int, identifier string) (*model.AuthIdentity, error)   // 根据登录方式和凭证获取登录认证
	GetAuthIdentityByIdentifier(ctx context.Context, identifier string) (*model.AuthIdentity, error)     // 根据凭证获取登录认证
	GetAuthIdentityByAuthType(ctx context.Context, uid int64, authType int) (*model.AuthIdentity, error) // 根据认证方式获取登录认证
	GetAuthIdentityByUID(ctx context.Context, uid int64) (string, string, error)                         // 获取用户身份认证
	GetPasswordHash(ctx context.Context, uid int64) (string, error)                                      // 根据 UID 获取用户密码
	UpdatePasswordHash(ctx context.Context, uid int64, passwordHash string) error                        // 修改用户密码
	HasPassword(ctx context.Context, uid int64) (bool, error)                                            // 查询密码状态
	SetPassword(ctx context.Context, authPassword *model.AuthPassword) error                             // 初始化密码
	DelRefreshToken(ctx context.Context, refreshToken string) error                                      // 缓存中删除 RefreshToken
	SetInfo(ctx context.Context, refreshToken string, mp map[string]any) error                           // 根据 RefreshToken 在缓存中存储用户信息
	GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error)          // 根据 RefreshToken 从缓存中读取用户信息
	SetBlackList(ctx context.Context, ssid string) error                                                 // 拉黑 SSID
	CheckBlackList(ctx context.Context, ssid string) (bool, error)                                       // 查看 SSID 是否被拉黑
}

type UserRepository interface {
	GetProfileByID(ctx context.Context, uid int64) (*model.UserProfile, error) // 根据 ID 查找用户资料
	UpdateProfile(ctx context.Context, id int64, updates map[string]any) error // 根据 ID 修改用户资料的多个字段
	Top(ctx context.Context) ([]*model.UserProfile, []float64, error)          // 返回热门推荐用户
	ChangeScore(ctx context.Context, uid int64, delta int) error               // 修改用户分数
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
	Retrieve(ctx context.Context, query string, scoreThreshold float64, limit int) ([]string, error)
	CreateChunksWithOutbox(ctx context.Context, chunkModels []*model.Chunk, event *model.Event) error
	UpsertVectorPoints(ctx context.Context, points []*qdrant.PointStruct) error
	GetChunksByBatchID(ctx context.Context, BatchID int64) ([]*model.Chunk, error)
}
