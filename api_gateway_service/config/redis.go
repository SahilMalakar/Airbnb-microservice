package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func LoadRedis() (*redis.Client, error) {
	otps, err := redis.ParseURL(RedisConfig.REDIS_URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %v", err)
	}

	rclient := redis.NewClient(otps)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rclient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %v", err)
	}

	fmt.Println("Successfully connected to redis")

	return rclient, nil
}
