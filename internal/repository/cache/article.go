package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lalalalade/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, articles []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error

	Set(ctx context.Context, art domain.Article) error
}

type RedisArticleCache struct {
	client redis.Cmdable
}

func NewRedisArticleCache(client redis.Cmdable) ArticleCache {
	return &RedisArticleCache{client: client}
}

func (r RedisArticleCache) GetFirstPage(ctx context.Context, author int64) ([]domain.Article, error) {
	bs, err := r.client.Get(ctx, r.firstPageKey(author)).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(bs, &arts)
	return arts, err
}

func (r RedisArticleCache) SetFirstPage(ctx context.Context, author int64, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	data, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.firstPageKey(author), data, time.Minute*10).Err()
}

func (r RedisArticleCache) firstPageKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}

func (r RedisArticleCache) key(id int64) string {
	return fmt.Sprintf("article:%d", id)
}

func (r RedisArticleCache) DelFirstPage(ctx context.Context, author int64) error {
	return r.client.Del(ctx, r.firstPageKey(author)).Err()
}

func (r RedisArticleCache) Set(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	// 过期时间要短
	return r.client.Set(ctx, r.key(art.Id), data, time.Second*10).Err()
}
