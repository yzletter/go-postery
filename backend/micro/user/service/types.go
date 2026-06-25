package service

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/micro/user/domain"
)

// UserService 定义用户微服务业务接口
type UserService interface {
	// GetProfile 获取用户资料
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- domain.Profile: 用户资料
	//	- error: 可能返回的错误
	//		- ErrInvalidArgument: 请求参数错误
	//		- ErrNotFound: 用户资料不存在
	//		- ErrInternal: 系统内部错误
	GetProfile(ctx context.Context, id int64) (domain.Profile, error)

	// UpdateProfile 更新用户资料
	//
	// Parameter:
	//	- id: 用户 ID
	//	- updates: 待更新字段
	//
	// Return:
	//	- error: 可能返回的错误
	//		- ErrInvalidArgument: 请求参数错误
	//		- ErrNotFound: 用户资料不存在
	//		- ErrAlreadyExits: 昵称已存在
	//		- ErrInternal: 系统内部错误
	UpdateProfile(ctx context.Context, id int64, updates map[string]any) error

	// UploadAvatarSign 获取上传头像 OSS 签名
	//
	// Parameter:
	//	- id: 用户 ID
	//
	// Return:
	//	- string: 上传签名
	//	- error: 可能返回的错误
	//		- ErrInvalidArgument: 请求参数错误
	//		- ErrInternal: OSS 签名生成失败
	UploadAvatarSign(ctx context.Context, id int64) (string, error)

	// UploadAvatarCallback 处理头像上传回调
	//
	// Parameter:
	//	- id: 用户 ID
	//	- object: 头像对象名称
	//
	// Return:
	//	- error: 可能返回的错误
	//		- ErrInvalidArgument: 请求参数错误
	//		- ErrNotFound: 用户资料不存在
	//		- ErrInternal: 系统内部错误
	UploadAvatarCallback(ctx context.Context, id int64, object string) error

	// GetAvatarURL 获取头像访问预签名 URL
	//
	// Parameter:
	//	- object: 头像对象名称
	//
	// Return:
	//	- string: 头像访问地址
	//	- error: 可能返回的错误
	//		- ErrInvalidArgument: 请求参数错误
	//		- ErrInternal: OSS 预签名 URL 生成失败
	GetAvatarURL(ctx context.Context, object string) (string, error)

	// GetIDAfterTime 根据时间查找之后创建的用户 ID
	//
	// Parameter:
	//	- timeAfter: 查询起始时间
	//
	// Return:
	//	- []int64: 用户 ID 列表
	//	- error: 可能返回的错误
	//		- ErrInvalidArgument: 请求参数错误
	//		- ErrInternal: 系统内部错误
	GetIDAfterTime(ctx context.Context, timeAfter time.Time) ([]int64, error)

	// ListFollowees 按页查找关注的人
	//
	// Parameter:
	//	- id: 用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 关注的人总数
	//	- []domain.ProfileBrief: 当前页关注的用户资料
	//	- error: 可能返回的错误
	//		- ErrInvalidArgument: 请求参数错误
	//		- ErrInternal: 系统内部错误
	ListFollowees(ctx context.Context, id int64, pageNo int, pageSize int) (int64, []domain.ProfileBrief, error)

	// ListFollowers 按页查找粉丝
	//
	// Parameter:
	//	- id: 用户 ID
	//	- pageNo: 页数
	//	- pageSize: 每页大小
	//
	// Return:
	//	- int64: 粉丝总数
	//	- []domain.ProfileBrief: 当前页的粉丝资料
	//	- error: 可能返回的错误
	//		- ErrInvalidArgument: 请求参数错误
	//		- ErrInternal: 系统内部错误
	ListFollowers(ctx context.Context, id int64, pageNo int, pageSize int) (int64, []domain.ProfileBrief, error)

	// Top 获取用户排行榜
	//
	// Return:
	//	- []domain.ProfileTop: 排行榜用户资料
	//	- error: 可能返回的错误
	//		- ErrInternal: 系统内部错误
	Top(ctx context.Context) ([]domain.ProfileTop, error)
}
