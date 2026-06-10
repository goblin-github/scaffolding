package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type Cache struct {
	core *redis.Client
}

func NewRedis(opt Options) (*Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     opt.Addr,
		Password: opt.Password,
		DB:       opt.DB,
		PoolSize: opt.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Cache{core: rdb}, nil
}

func (r *Cache) Get(ctx context.Context, key string) (string, error) {
	return r.core.Get(ctx, key).Result()
}

func (r *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.core.Set(ctx, key, value, ttl).Err()
}

func (r *Cache) Del(ctx context.Context, keys ...string) error {
	return r.core.Del(ctx, keys...).Err()
}

func (r *Cache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.core.Exists(ctx, keys...).Result()
}

func (r *Cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.core.Expire(ctx, key, ttl).Err()
}

func (r *Cache) Incr(ctx context.Context, key string) (int64, error) {
	return r.core.Incr(ctx, key).Result()
}

func (r *Cache) HGet(ctx context.Context, key, field string) (string, error) {
	return r.core.HGet(ctx, key, field).Result()
}

func (r *Cache) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.core.HSet(ctx, key, values...).Err()
}

func (r *Cache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.core.HGetAll(ctx, key).Result()
}

func (r *Cache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return r.core.SetNX(ctx, key, value, ttl).Result()
}

// Native exposes the underlying redis.Client for use cases
// the wrapper doesn't cover (e.g. pipelines, custom commands).
func (r *Cache) Native() *redis.Client {
	return r.core
}

func (r *Cache) Close() error {
	if r == nil {
		return nil
	}
	return r.core.Close()
}
