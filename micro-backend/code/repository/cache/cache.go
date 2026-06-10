package cache

import (
	"context"
	_ "embed"
	"errors"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/micro-backend/code/conf"
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

	res, err := cache.client.Eval(ctx, verifyScript, []string{key}, code).Int()
	if err != nil {
		return false, err
	}

	switch res {
	case 0: // 验证码不存在或已过期
		return false, nil
	case 1: // 验证码错误
		return false, nil
	case 2: // 验证码正确
		return true, nil
	default:
		slog.Error("Unexpected Error")
		return false, errors.New("Server Internal")
	}

	//if ok, err := cache.client.Eval(ctx, verifyScript, []string{key}, code).Bool(); err != nil {
	//	slog.Error("Redis Eval Script Failed", "error", err.Error())
	//	return false, err
	//} else {
	//	return ok, nil
	//}
}
