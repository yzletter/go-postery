package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/backend/micro/interview/model"
)

const (
	profileKeyPrefix = "interview:profile:"
	sessionKeyPrefix = "interview:session:"

	ProfileExpireTime = 10 * time.Minute
	SessionExpireTime = 30 * time.Minute
)

// redisInterviewCache 用 Redis 实现 InterviewCache
type redisInterviewCache struct {
	client redis.UniversalClient
}

// NewInterviewCache 构造函数
func NewInterviewCache(client redis.UniversalClient) InterviewCache {
	return &redisInterviewCache{client: client}
}

// SetProfile 设置用户画像缓存
func (cache *redisInterviewCache) SetProfile(ctx context.Context, profile *model.InterviewProfile) error {
	// 0. 兜底
	if profile == nil || profile.UserID == 0 {
		return ErrParamsInvalid
	}

	// 1. 序列化
	bytes, err := sonic.Marshal(profile)
	if err != nil {
		return ErrServerInternal
	}

	// 2. 写入 Redis
	if err := cache.client.Set(ctx, newProfileKey(profile.UserID), bytes, ProfileExpireTime).Err(); err != nil {
		return ErrServerInternal
	}
	return nil
}

// GetProfile 获取用户画像缓存
func (cache *redisInterviewCache) GetProfile(ctx context.Context, userID int64) (*model.InterviewProfile, error) {
	// 0. 兜底
	if userID == 0 {
		return nil, ErrParamsInvalid
	}

	// 1. 读取 Redis
	bytes, err := cache.client.Get(ctx, newProfileKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}

	// 2. 反序列化
	var profile model.InterviewProfile
	if err := sonic.Unmarshal(bytes, &profile); err != nil {
		return nil, ErrServerInternal
	}
	return &profile, nil
}

// DelProfile 删除用户画像缓存
func (cache *redisInterviewCache) DelProfile(ctx context.Context, userID int64) error {
	if userID == 0 {
		return ErrParamsInvalid
	}
	return cache.client.Del(ctx, newProfileKey(userID)).Err()
}

// SetSession 设置面试会话缓存
func (cache *redisInterviewCache) SetSession(ctx context.Context, sessionID int64, data []byte) error {
	// 0. 兜底
	if sessionID == 0 {
		return ErrParamsInvalid
	}

	// 1. 写入 Redis
	if err := cache.client.Set(ctx, newSessionKey(sessionID), data, SessionExpireTime).Err(); err != nil {
		return ErrServerInternal
	}
	return nil
}

// GetSession 获取面试会话缓存
func (cache *redisInterviewCache) GetSession(ctx context.Context, sessionID int64) ([]byte, error) {
	if sessionID == 0 {
		return nil, ErrParamsInvalid
	}
	return cache.client.Get(ctx, newSessionKey(sessionID)).Bytes()
}

// DelSession 删除面试会话缓存
func (cache *redisInterviewCache) DelSession(ctx context.Context, sessionID int64) error {
	if sessionID == 0 {
		return ErrParamsInvalid
	}
	return cache.client.Del(ctx, newSessionKey(sessionID)).Err()
}

// newProfileKey 构造用户画像缓存 Key
func newProfileKey(userID int64) string {
	return profileKeyPrefix + strconv.FormatInt(userID, 10)
}

// newSessionKey 构造面试会话缓存 Key
func newSessionKey(sessionID int64) string {
	return sessionKeyPrefix + strconv.FormatInt(sessionID, 10)
}
