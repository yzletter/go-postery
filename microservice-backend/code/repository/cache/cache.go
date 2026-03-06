package cache

import (
	"context"
	_ "embed"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/microservice-backend/code/conf"
)

//go:embed lua/allow.lua
var allowScript string // 用于检查发送验证码的 Lua 脚本

//go:embed lua/verify.lua
var verifyScript string // 用于校验验证码的 Lua 脚本

type redisCodeCache struct {
	client redis.UniversalClient
}

func NewCodeCache(client redis.UniversalClient) CodeCache {
	return &redisCodeCache{client: client}
}

// Allow 检查发送验证码
func (cache *redisCodeCache) Allow(ctx context.Context, biz int, identifier string, code string) (int, error) {
	var interval int
	var key string
	var validTime int

	switch biz {
	case 1:
		key = conf.PhoneCodePrefix + identifier
		interval = conf.SendSMSInterval
		validTime = conf.SMSValidTime
	case 2:
		key = conf.EmailCodePrefix + identifier
		interval = conf.SendEmailInterval
		validTime = conf.EmailValidTime
	default:
		return -1, nil
	}

	result, err := cache.client.Eval(ctx, allowScript, []string{key}, code, interval, validTime).Int()
	return result, err
}

// Verify 校验验证码
func (cache *redisCodeCache) Verify(ctx context.Context, biz int, identifier string, code string) (bool, error) {
	var key string
	switch biz {
	case 1:
		key = conf.PhoneCodePrefix + identifier
	case 2:
		key = conf.EmailCodePrefix + identifier
	default:
		return false, nil
	}

	if ok, err := cache.client.Eval(ctx, verifyScript, []string{key}, code).Bool(); err != nil {
		return false, err
	} else {
		return ok, nil
	}
}
