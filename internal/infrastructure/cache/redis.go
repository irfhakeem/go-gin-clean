package cache

import (
	"context"
	"encoding/json"
	"time"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/pkg/config"
	pkgerrors "go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/message"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	cfg    *config.RedisConfig
	client *redis.Client
}

func NewRedisCache(client *redis.Client, cfg *config.RedisConfig) port.Cache {
	return &redisCache{cfg: cfg, client: client}
}

func (r *redisCache) Set(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	exp := time.Duration(r.cfg.Expiration) * time.Second
	return r.client.Set(ctx, key, data, exp).Err()
}

func (r *redisCache) Get(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return pkgerrors.WrapAppError(pkgerrors.NotFound, message.ErrCacheMiss, pkgerrors.ErrCacheMiss)
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (r *redisCache) GetString(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	res, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}

func (r *redisCache) SetWithExpiration(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}
