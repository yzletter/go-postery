package dao

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
	"github.com/yzletter/go-postery/dto/session"
	"github.com/yzletter/go-postery/model"
)

// 定义 DAO 层所有接口

type AuthDAO interface {
	CreateUser(ctx context.Context, authAggregate *model.AuthAggregate) error                            // 创建用户（包括用户最小项、用户登录认证、用户密码、用户资料、注册扩展功能）
	GetAuthIdentity(ctx context.Context, authType int, identifier string) (*model.AuthIdentity, error)   // 根据登录方式和凭证获取登录认证
	GetAuthIdentityByIdentifier(ctx context.Context, identifier string) (*model.AuthIdentity, error)     // 根据凭证获取登录认证
	GetAuthIdentityByAuthType(ctx context.Context, uid int64, authType int) (*model.AuthIdentity, error) // 根据认证方式获取登录认证
	GetAuthIdentityByUID(ctx context.Context, uid int64) ([]*model.AuthIdentity, error)                  // 获取用户身份认证
	GetPasswordHash(ctx context.Context, uid int64) (string, error)                                      // 根据 UID 获取用户密码
	UpdatePasswordHash(ctx context.Context, uid int64, passwordHash string) error                        // 修改用户密码
	HasPassword(ctx context.Context, uid int64) (bool, error)                                            // 查询密码状态
	SetPassword(ctx context.Context, authPassword *model.AuthPassword) error                             // 初始化密码
}

type UserDAO interface {
	GetProfileByID(ctx context.Context, id int64) (*model.UserProfile, error)  // 根据 ID 查找用户资料
	UpdateProfile(ctx context.Context, id int64, updates map[string]any) error // 根据 ID 修改用户资料的多个字段
}

type PostDAO interface {
	Create(ctx context.Context, post *model.Post, events []*model.Event) error
	Delete(ctx context.Context, id int64) error
	UpdateCount(ctx context.Context, id int64, field model.PostCntField, delta int) error
	Update(ctx context.Context, id int64, updates map[string]any) error
	GetByID(ctx context.Context, id int64) (*model.Post, error)
	GetByUid(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Post, error)
	GetByPage(ctx context.Context, pageNo, pageSize int) (int64, []*model.Post, error)
	GetByPageAndTag(ctx context.Context, tid int64, pageNo, pageSize int) (int64, []*model.Post, error)
}

type CommentDAO interface {
	Create(ctx context.Context, comment *model.Comment) error
	Delete(ctx context.Context, id int64) (int, error)
	GetByID(ctx context.Context, id int64) (*model.Comment, error)
	GetByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Comment, error)
	GetRepliesByParentID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Comment, error)
}

type LikeDAO interface {
	Create(ctx context.Context, like *model.Like) error
	Delete(ctx context.Context, uid, pid int64) error
	Exists(ctx context.Context, uid, pid int64) (bool, error)
}

type FollowDAO interface {
	Create(ctx context.Context, follow *model.Follow) error
	Delete(ctx context.Context, ferID, feeID int64) error
	Exists(ctx context.Context, ferID, feeID int64) (model.FollowType, error)
	GetFollowers(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)
	GetFollowees(ctx context.Context, id int64, pageNo, pageSize int) (int64, []int64, error)
}

type TagDAO interface {
	Create(ctx context.Context, tag *model.Tag) error
	GetBySlug(ctx context.Context, slug string) (*model.Tag, error)
	GetByName(ctx context.Context, name string) (*model.Tag, error)
	Bind(ctx context.Context, postTag *model.PostTag) error
	DeleteBind(ctx context.Context, pid, tid int64) error
	FindTagsByPostID(ctx context.Context, pid int64) ([]string, error)
}

type MessageDAO interface {
	Create(ctx context.Context, message *model.Message) error
	GetByIDAndTargetID(ctx context.Context, id, targetID int64) ([]*model.Message, error)
	GetByPage(ctx context.Context, id int64, targetID int64, pageNo, pageSize int) (int64, []*model.Message, error)
}

type SessionDAO interface {
	Create(ctx context.Context, session *model.Session) error
	GetByUid(ctx context.Context, uid int64) ([]*model.Session, error)
	GetByUidAndTargetID(ctx context.Context, uid, targetID int64) (*model.Session, error)
	GetByID(ctx context.Context, uid, sid int64) (*model.Session, error)
	Delete(ctx context.Context, uid, sid int64) error
	UpdateUnread(ctx context.Context, uid int64, sid int64, updates session.UpdateUnreadRequest) error
	ClearUnread(ctx context.Context, uid int64, sid int64) error
}

type SmsDAO interface {
}

type OrderDAO interface {
	Create(ctx context.Context, order *model.Order) error
	Get(ctx context.Context, uid int64) (*model.Order, error)
}

type GiftDAO interface {
	GetAll(ctx context.Context) ([]*model.Gift, error)
	GetByID(ctx context.Context, gid int64) (*model.Gift, error)
}

type AgentDAO interface {
	Retrieve(ctx context.Context, query string, scoreThreshold float64, limit int) ([]string, error)
	CreateChunksWithOutbox(ctx context.Context, chunkModels []*model.Chunk, event *model.Event) error
	UpsertVectorPoints(ctx context.Context, points []*qdrant.PointStruct) error
	GetChunksByBatchID(ctx context.Context, BatchID int64) ([]*model.Chunk, error)
}
