package cache

import (
	"context"
	_ "embed"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/conf"
)

const (
	phoneCodePrefix = "phone:code:"
)

type redisSmsCache struct {
	client redis.UniversalClient
}

//go:embed lua/check_sms_code.lua
var checkCodeScript string

func NewSmsCache(client redis.UniversalClient) SmsCache {
	return &redisSmsCache{client: client}
}

func (cache *redisSmsCache) CheckCode(ctx context.Context, phoneNumber string, code string) (int, error) {
	key := phoneCodePrefix + phoneNumber
	result, err := cache.client.Eval(ctx, checkCodeScript, []string{key}, code, conf.SendSMSInterval, conf.SMSValidTime).Int()
	return result, err
}
