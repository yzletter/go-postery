package repository

import (
	_ "embed"
	"fmt"

	"github.com/go-redis/redis"
)

const (
	POST_IN_REDIS = "post:"
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
	field := "view_cnt"

	return repo.redisClient.Eval(addCntScript, []string{redisKey}, field, delta).Bool()
}

func (repo *PostCacheRepository) ChangeLikeCnt(pid int, delta int) (bool, error) {
	redisKey := fmt.Sprintf("%s:%d", POST_IN_REDIS, pid)
	field := "like_cnt"

	return repo.redisClient.Eval(addCntScript, []string{redisKey}, field, delta).Bool()
}

func (repo *PostCacheRepository) ChangeCommentCnt(pid int, delta int) (bool, error) {
	redisKey := fmt.Sprintf("%s:%d", POST_IN_REDIS, pid)
	field := "comment_cnt"

	return repo.redisClient.Eval(addCntScript, []string{redisKey}, field, delta).Bool()
}

func (repo *PostCacheRepository) SetKey(pid int, filed string, val int) {
	redisKey := fmt.Sprintf("%s:%d", POST_IN_REDIS, pid)
	repo.redisClient.HSet(redisKey, filed, val)
}
