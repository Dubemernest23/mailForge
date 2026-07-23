package redisclient

import (
	"context"
	"mailForgeApi/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {

	url := cfg.Redis
	opt, err := redis.ParseURL(url.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return client, nil
}
