package repository

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const (
	POST_IN_REDIS    = "post_interactive_prefix"
	POST_EXPIRE_TIME = time.Minute * 15
)

//go:embed add_cnt_script.lua
var addCntScript string

type PostCacheRepository struct {
	redisClient redis.Cmdable
}

func NewPostCacheRepository(redisClient redis.Cmdable) *PostCacheRepository {
	return &PostCacheRepository{
		redisClient: redisClient,
	}
}

func (repo *PostCacheRepository) ChangeViewCnt(pid int, delta int) (bool, error) {
	redisKey := fmt.Sprintf("%s:%d", POST_IN_REDIS, pid)
	field := "view_count"

	return repo.redisClient.Eval(addCntScript, []string{redisKey}, field, delta).Bool()
}

func (repo *PostCacheRepository) ChangeLikeCnt(pid int, delta int) (bool, error) {
	redisKey := fmt.Sprintf("%s:%d", POST_IN_REDIS, pid)
	field := "like_count"

	return repo.redisClient.Eval(addCntScript, []string{redisKey}, field, delta).Bool()
}

func (repo *PostCacheRepository) ChangeCommentCnt(pid int, delta int) (bool, error) {
	redisKey := fmt.Sprintf("%s:%d", POST_IN_REDIS, pid)
	field := "comment_count"

	return repo.redisClient.Eval(addCntScript, []string{redisKey}, field, delta).Bool()
}

func (repo *PostCacheRepository) SetKey(pid int, fields []string, vals []int) {
	// 拼接 Key
	redisKey := fmt.Sprintf("%s:%d", POST_IN_REDIS, pid)
	// 设置 Key
	repo.redisClient.HMSet(redisKey,
		map[string]interface{}{
			fields[0]: vals[0],
			fields[1]: vals[1],
			fields[2]: vals[2],
		})
	// 设置过期时间
	repo.redisClient.Expire(redisKey, POST_EXPIRE_TIME)
}
