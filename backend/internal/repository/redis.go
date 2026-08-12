package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"sanoc/backend/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func InitRedis(cfg *config.Config) (*RedisClient, error) {
	addr := cfg.RedisAddr()
	log.Printf("[Redis] Connecting to Redis cache server at %s...", addr)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping redis server at %s: %w", addr, err)
	}

	log.Printf("[Redis] PING success (%s) — Live device status cache & send queue ready", pong)
	return &RedisClient{Client: rdb}, nil
}
