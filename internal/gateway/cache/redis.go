package cache

import (
	"context"
	"encoding/json"
	"go-gin-clean/pkg/config"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	cfg    *config.RedisConfig
	client *redis.Client
}

func NewRedisService(cfg *config.RedisConfig) *RedisService {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + strconv.Itoa(cfg.Port), // Addr biasanya berupa "host:port", misal "localhost:6379"
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &RedisService{cfg: cfg, client: client}
}

func (r *RedisService) Set(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	exp := time.Duration(r.cfg.Expiration) * time.Second
	return r.client.Set(ctx, key, data, exp).Err()
}

func (r *RedisService) Get(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return redis.Nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (r *RedisService) GetString(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisService) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisService) DeletePattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (r *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	res, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}

func (r *RedisService) SetWithExpiration(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}
