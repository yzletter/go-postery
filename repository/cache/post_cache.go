package cache

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/model"
)

const (
	postInteractiveKeyPrefix = "post:interactive"
	POST_EXPIRE_TIME         = time.Minute * 15
)

//go:embed lua/change_cnt_script.lua
var addCntScript string

type redisPostCache struct {
	client redis.UniversalClient
}

func NewPostCache(client redis.UniversalClient) PostCache {
	return &redisPostCache{client: client}
}

// ChangeInteractiveCnt HIncrBy KEY 对应 Field 的值, 值为 Delta
func (cache *redisPostCache) ChangeInteractiveCnt(ctx context.Context, pid int64, field model.PostCntField, delta int) (bool, error) {
	col, err := field.Column()
	if err != nil {
		return false, err
	}
	redisKey := fmt.Sprintf("%s:%d", postInteractiveKeyPrefix, pid)
	return cache.client.Eval(ctx, addCntScript, []string{redisKey}, col, delta).Bool()
}

func (cache *redisPostCache) SetKey(ctx context.Context, pid int64, fields []model.PostCntField, vals []int) {
	// 拼接 Key
	redisKey := fmt.Sprintf("%s:%d", postInteractiveKeyPrefix, pid)

	// 设置 Key
	mp := make(map[string]interface{})
	for k, v := range fields {
		s, _ := v.Column()
		mp[s] = vals[k]
	}
	cache.client.HSet(ctx, redisKey, mp)

	// 设置过期时间
	cache.client.Expire(ctx, redisKey, POST_EXPIRE_TIME)
}
