package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/backend/micro/user/domain"
)

const (
	profileKeyPrefix     = "user:profile:"  // 缓存 Key 的前缀
	profileKeyExpireTime = 10 * time.Minute // 缓存过期时间
)

// redisUserCache 用 Redis 实现 UserCache
type redisUserCache struct {
	client redis.UniversalClient
}

// NewUserCache 构造函数
func NewUserCache(client redis.UniversalClient) UserCache {
	return &redisUserCache{client: client}
}

// GetProfile 获取用户资料缓存
func (cache *redisUserCache) GetProfile(ctx context.Context, id int64) (domain.Profile, error) {
	var profile domain.Profile
	if err := cache.get(ctx, newProfileKey(id), &profile); err != nil {
		return domain.Profile{}, err
	}
	return profile, nil
}

// SetProfile 设置用户资料缓存
func (cache *redisUserCache) SetProfile(ctx context.Context, id int64, profile domain.Profile) error {
	return cache.set(ctx, newProfileKey(id), profile, profileKeyExpireTime)
}

// DelProfile 删除用户资料缓存
func (cache *redisUserCache) DelProfile(ctx context.Context, id int64) error {
	return cache.client.Del(ctx, newProfileKey(id)).Err()
}

// newProfileKey 构造用户资料缓存 Key
func newProfileKey(id int64) string {
	return profileKeyPrefix + strconv.FormatInt(id, 10)
}

// set 序列化数据并写入缓存
func (cache *redisUserCache) set(ctx context.Context, key string, val any, expiration time.Duration) error {
	// 序列化
	bytes, err := sonic.Marshal(val)
	if err != nil {
		return ErrServerInternal
	}

	// 写入 Redis
	if err := cache.client.Set(ctx, key, bytes, expiration).Err(); err != nil {
		return ErrServerInternal
	}
	return nil
}

// get 从缓存读取数据并反序列化
func (cache *redisUserCache) get(ctx context.Context, key string, dst any) error {
	// 从 Redis 读取数据
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	// 反序列化
	if err := sonic.Unmarshal(val, dst); err != nil {
		return ErrServerInternal
	}
	return nil
}
