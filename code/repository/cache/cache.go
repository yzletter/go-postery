package cache

import (
	"context"
	_ "embed"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/code/conf"
	"github.com/yzletter/go-postery/code/model"
)

//go:embed lua/allow_send_code.lua
var allowSendCodeScript string

//go:embed lua/check_code.lua
var checkCodeScript string

type redisCodeCache struct {
	client redis.UniversalClient
}

func NewCodeCache(client redis.UniversalClient) CodeCache {
	return &redisCodeCache{client: client}
}

// Allow 是否允许发送 Code
func (cache *redisCodeCache) Allow(ctx context.Context, biz model.CodeBiz, field string, code string) (int, error) {
	var interval int
	var key string
	var validTime int

	switch biz {
	case model.EmailCode:
		key = conf.EmailCodePrefix + field
		interval = conf.SendEmailInterval
		validTime = conf.EmailValidTime
	case model.SMSCode:
		key = conf.PhoneCodePrefix + field
		interval = conf.SendSMSInterval
		validTime = conf.SMSValidTime
	default:
		return -1, nil
	}

	result, err := cache.client.Eval(ctx, allowSendCodeScript, []string{key}, code, interval, validTime).Int()
	return result, err
}

func (cache *redisCodeCache) CheckCode(ctx context.Context, biz model.CodeBiz, field string, code string) (bool, error) {
	var key string
	switch biz {
	case model.EmailCode:
		key = conf.EmailCodePrefix + field
	case model.SMSCode:
		key = conf.PhoneCodePrefix + field
	default:
		return false, nil
	}

	ok, err := cache.client.Eval(ctx, checkCodeScript, []string{key}, code).Bool()
	if err != nil {
		return false, err
	}
	return ok, nil
}
