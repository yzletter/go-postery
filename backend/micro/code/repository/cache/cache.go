package cache

import (
	"context"
	_ "embed"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/micro/code/domain"
	"github.com/yzletter/go-postery/backend/micro/code/script"
)

type redisCodeCache struct {
	client redis.UniversalClient
}

func NewCodeCache(client redis.UniversalClient) CodeCache {
	return &redisCodeCache{client: client}
}

// Allow 检查发送验证码
func (cache *redisCodeCache) Allow(ctx context.Context, biz domain.BizType, identifier string, code string) (int, error) {
	var interval int
	var key string
	var validTime int

	switch biz {
	case domain.BizSMS:
		key = conf.PhoneCodePrefix + identifier
		interval = conf.SendSMSInterval
		validTime = conf.SMSValidTime
	case domain.BizEmail:
		key = conf.EmailCodePrefix + identifier
		interval = conf.SendEmailInterval
		validTime = conf.EmailValidTime
	default:
		return -1, nil
	}

	res, err := cache.client.Eval(ctx, script.AllowCodeScript, []string{key}, code, interval, validTime).Int()
	return res, err
}

// Verify 校验验证码
func (cache *redisCodeCache) Verify(ctx context.Context, biz domain.BizType, identifier string, code string) (int, error) {
	var key string
	switch biz {
	case domain.BizSMS:
		key = conf.PhoneCodePrefix + identifier
	case domain.BizEmail:
		key = conf.EmailCodePrefix + identifier
	default:
		return -1, nil
	}

	res, err := cache.client.Eval(ctx, script.VerifyCodeScript, []string{key}, code).Int()
	return res, err

}
