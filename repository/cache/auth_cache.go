package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/conf"
)

type redisAuthCache struct {
	client redis.UniversalClient
}

func (cache *redisAuthCache) GetPhoneCode(ctx context.Context, phone string) (string, error) {
	return cache.client.Get(ctx, conf.PhoneCodePrefix+phone).Result()
}

func (cache *redisAuthCache) GetEmailCode(ctx context.Context, email string) (string, error) {
	return cache.client.Get(ctx, conf.EmailCodePrefix+email).Result()
}

func (cache *redisAuthCache) DelRefreshToken(ctx context.Context, refreshToken string) error {
	return cache.client.Del(ctx, conf.RefreshTokenPrefix+refreshToken).Err()
}

func (cache *redisAuthCache) CheckBlackList(ctx context.Context, ssid string) (bool, error) {
	cnt, err := cache.client.Exists(ctx, conf.ClearTokenPrefix+ssid).Result()
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
}

func (cache *redisAuthCache) GetInfoByRefreshToken(ctx context.Context, refreshToken string) (int64, int, string, error) {
	mp, err := cache.client.HGetAll(ctx, conf.RefreshTokenPrefix+refreshToken).Result()
	if err != nil || len(mp) == 0 {
		return 0, 0, "", err
	}

	uid, err1 := strconv.ParseInt(mp["user_id"], 10, 64)
	role, err2 := strconv.Atoi(mp["role"])
	ssid := mp["ssid"]
	if ssid == "" || err1 != nil || err2 != nil {
		return 0, 0, "", err
	}

	return uid, role, ssid, nil
}

func (cache *redisAuthCache) SetBlackList(ctx context.Context, ssid string) error {
	ttl := time.Duration(conf.RefreshTokenMaxAgeSecs) * time.Second
	return cache.client.Set(ctx, conf.ClearTokenPrefix+ssid, "", ttl).Err()
}

func (cache *redisAuthCache) SetInfo(ctx context.Context, refreshToken string, mp map[string]any) error {
	ttl := time.Duration(conf.RefreshTokenMaxAgeSecs) * time.Second

	// 设置 Key Value
	err := cache.client.HSet(ctx, conf.RefreshTokenPrefix+refreshToken, mp).Err()
	if err != nil {
		return err
	}

	// 设置过期
	err = cache.client.Expire(ctx, conf.RefreshTokenPrefix+refreshToken, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

func NewAuthCache(client redis.UniversalClient) AuthCache {
	return &redisAuthCache{
		client: client,
	}
}
