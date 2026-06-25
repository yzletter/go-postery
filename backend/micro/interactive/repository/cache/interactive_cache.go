package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"github.com/yzletter/go-postery/backend/micro/interactive/domain"
)

const (
	postInteractiveKeyPrefix = "interactive:post:"
	userInteractiveKeyPrefix = "interactive:user:"
	likeKeyPrefix            = "interactive:like:"
	followKeyPrefix          = "interactive:follow:"
	consumeKeyPrefix         = "interactive:consume:"
	interactiveExpireTime    = 3 * time.Minute
)

type redisInteractiveCache struct {
	client redis.UniversalClient
}

func (cache *redisInteractiveCache) GetConsume(ctx context.Context, consumer string, id int64) (bool, error) {
	res, err := cache.client.Exists(ctx, consumeKey(consumer, id)).Result()
	if err != nil {
		return false, err
	}
	if res == 0 {
		return false, redis.Nil
	}
	return res > 0, nil
}

func (cache *redisInteractiveCache) SetConsume(ctx context.Context, consumer string, id int64) error {
	return cache.client.Set(ctx, consumeKey(consumer, id), 0, interactiveExpireTime).Err()
}

func NewInteractiveCache(client redis.UniversalClient) InteractiveCache {
	return &redisInteractiveCache{client: client}
}

func (cache *redisInteractiveCache) GetPostInteractive(ctx context.Context, id int64) (domain.PostInter, error) {
	var inter domain.PostInter
	if err := cache.getJSON(ctx, postInteractiveKey(id), &inter); err != nil {
		return domain.PostInter{}, err
	}
	return inter, nil
}

func (cache *redisInteractiveCache) SetPostInteractive(ctx context.Context, id int64, inter domain.PostInter) error {
	return cache.setJSON(ctx, postInteractiveKey(id), inter)
}

func (cache *redisInteractiveCache) DelPostInteractive(ctx context.Context, id int64) error {
	return cache.client.Del(ctx, postInteractiveKey(id)).Err()
}

func (cache *redisInteractiveCache) GetUserInteractive(ctx context.Context, id int64) (domain.UserInter, error) {
	var inter domain.UserInter
	if err := cache.getJSON(ctx, userInteractiveKey(id), &inter); err != nil {
		return domain.UserInter{}, err
	}
	return inter, nil
}

func (cache *redisInteractiveCache) SetUserInteractive(ctx context.Context, id int64, inter domain.UserInter) error {
	return cache.setJSON(ctx, userInteractiveKey(id), inter)
}

// DelUserInteractive 删除用户 Inter 缓存
func (cache *redisInteractiveCache) DelUserInteractive(ctx context.Context, id int64) error {
	return cache.client.Del(ctx, userInteractiveKey(id)).Err()
}

func (cache *redisInteractiveCache) GetLike(ctx context.Context, uid, pid int64) (bool, error) {
	res, err := cache.client.Exists(ctx, likeKey(uid, pid)).Result()
	if err != nil {
		return false, err
	}
	if res == 0 {
		return false, redis.Nil
	}
	return res > 0, nil
}

func (cache *redisInteractiveCache) SetLike(ctx context.Context, uid, pid int64) error {
	return cache.client.Set(ctx, likeKey(uid, pid), 1, interactiveExpireTime).Err()
}

func (cache *redisInteractiveCache) DelLike(ctx context.Context, uid, pid int64) error {
	return cache.client.Del(ctx, likeKey(uid, pid)).Err()
}

func (cache *redisInteractiveCache) GetFollow(ctx context.Context, follower, followee int64) (domain.FollowType, error) {
	keys := []string{followKey(follower, followee), followKey(followee, follower)}
	cnt, err := cache.client.Exists(ctx, keys...).Result()
	if err != nil {
		return domain.FollowWrong, err
	}

	switch cnt {
	case 2:
		// 两个 Key 都有
		return domain.FollowMutual, nil
	case 1:
		// 只有一个 Key
		ok, err := cache.client.Exists(ctx, keys[0]).Result()
		if err != nil {
			return domain.FollowWrong, err
		}
		if ok > 0 {
			return domain.FollowIFollow, nil
		}
		return domain.FollowFollowMe, nil
	default:
		return domain.FollowNone, redis.Nil
	}
}

func (cache *redisInteractiveCache) SetFollow(ctx context.Context, follower, followee int64) error {
	return cache.client.Set(ctx, followKey(follower, followee), "", interactiveExpireTime).Err()
}

func (cache *redisInteractiveCache) DelFollow(ctx context.Context, follower, followee int64) error {
	return cache.client.Del(ctx, followKey(follower, followee)).Err()
}

func (cache *redisInteractiveCache) getJSON(ctx context.Context, key string, dst any) error {
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return sonic.Unmarshal(val, dst)
}

func (cache *redisInteractiveCache) setJSON(ctx context.Context, key string, val any) error {
	data, err := sonic.Marshal(val)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, key, data, interactiveExpireTime).Err()
}

func postInteractiveKey(id int64) string {
	return postInteractiveKeyPrefix + strconv.FormatInt(id, 10)
}

func userInteractiveKey(id int64) string {
	return userInteractiveKeyPrefix + strconv.FormatInt(id, 10)
}

func likeKey(uid, pid int64) string {
	return fmt.Sprintf("%s%d:%d", likeKeyPrefix, uid, pid)
}

func followKey(follower, followee int64) string {
	return fmt.Sprintf("%s%d:%d", followKeyPrefix, follower, followee)
}

func consumeKey(consumer string, id int64) string {
	return fmt.Sprintf("%s%s:%d", consumeKeyPrefix, consumer, id)
}
