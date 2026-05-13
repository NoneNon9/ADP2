package repository

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisIdempotencyStore struct {
	client *redis.Client
}

func NewRedisIdempotencyStore(redisURL string) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{
		client: redis.NewClient(&redis.Options{Addr: redisURL}),
	}
}

func (s *RedisIdempotencyStore) MarkAsProcessing(ctx context.Context, eventID string) bool {

	success, _ := s.client.SetNX(ctx, "processed_event:"+eventID, "done", 24*time.Hour).Result()
	return success
}
