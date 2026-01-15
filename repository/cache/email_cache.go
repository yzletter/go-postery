package cache

import (
	"context"
	_ "embed"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/conf"
)

const ()

//go:embed lua/check_sms_code.lua
var checkEmailCodeScript string

type redisEmailCache struct {
	client redis.UniversalClient
}

func NewEmailCache(client redis.UniversalClient) EmailCache {
	return &redisEmailCache{
		client: client,
	}
}

func (cache *redisEmailCache) CheckCode(ctx context.Context, emailAddress string, code string) (int, error) {
	key := conf.EmailCodePrefix + emailAddress
	result, err := cache.client.Eval(ctx, checkEmailCodeScript, []string{key}, code, conf.SendEmailInterval, conf.EmailValidTime).Int()
	return result, err
}
