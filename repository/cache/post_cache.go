package cache

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
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

func (cache *redisPostCache) DeleteScore(ctx context.Context, id int64) error {
	_, err := cache.client.ZRem(ctx, model.KeyPostScore, id).Result()
	if err != nil {
		return err
	}
	_, err = cache.client.ZRem(ctx, model.KeyPostTime, id).Result()
	if err != nil {
		return err
	}
	return err
}

func (cache *redisPostCache) Top(ctx context.Context) ([]int64, []float64, error) {
	pairs, err := cache.client.ZRevRangeWithScores(ctx, model.KeyPostScore, 0, 9).Result()
	if err != nil {
		return nil, nil, err
	}
	var ids []int64
	var scores []float64
	for _, pair := range pairs {
		id, err := strconv.ParseInt(pair.Member.(string), 10, 64)
		if err != nil {
			id = 0
		}
		score := pair.Score
		ids = append(ids, id)
		scores = append(scores, score)
	}

	return ids, scores, nil
}

func (cache *redisPostCache) ChangeScore(ctx context.Context, pid int64, delta int) error {
	_, err := cache.client.ZIncrBy(ctx, model.KeyPostScore, float64(delta), strconv.FormatInt(pid, 10)).Result()
	return err
}

func (cache *redisPostCache) CheckPostLikeTime(ctx context.Context, pid int64) (float64, error) {
	return cache.client.ZScore(ctx, model.KeyPostTime, strconv.Itoa(int(pid))).Result()
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

// SetInteractiveKey 设置 Post 的 浏览、点赞、评论 Key
func (cache *redisPostCache) SetInteractiveKey(ctx context.Context, pid int64, fields []model.PostCntField, vals []int) {
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

// SetScore 设置初始分数
func (cache *redisPostCache) SetScore(ctx context.Context, pid int64) error {
	_, err := cache.client.ZAdd(ctx, model.KeyPostScore, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: pid,
	}).Result()
	if err != nil {
		return err
	}

	_, err = cache.client.ZAdd(ctx, model.KeyPostTime, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: pid,
	}).Result()
	if err != nil {
		return err
	}

	return nil
}
