package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lalalalade/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type UserCache struct {
	// 传单机 Redis 可以
	// 传 cluster 的 Redis 也可以
	client     redis.Cmdable
	expiration time.Duration
}

func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

// Get 只要 error 为 nil，就认为缓存一定有数据
// 如果没有数据，返回一个特定的 error
func (cache *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	// 数据不存在 err 为 redis.Nil
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *UserCache) Set(ctx context.Context, user domain.User) error {
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}
	key := cache.key(user.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *UserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
