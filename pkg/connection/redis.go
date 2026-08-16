package connection

import (
	"context"
	"strconv"

	"go-gin-clean/pkg/config"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

func CloseRedis(client *redis.Client) error {
	if client == nil {
		return nil
	}
	return client.Close()
}
