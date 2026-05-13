package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"order-service/internal/domain"
)

type RedisOrderCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisOrderCache(redisURL string) *RedisOrderCache {
	client := redis.NewClient(&redis.Options{Addr: redisURL})
	return &RedisOrderCache{
		client: client,
		ttl:    5 * time.Minute,
	}
}

func (c *RedisOrderCache) Set(ctx context.Context, order domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "order:"+order.ID, data, c.ttl).Err()
}

func (c *RedisOrderCache) Get(ctx context.Context, id string) (domain.Order, error) {
	val, err := c.client.Get(ctx, "order:"+id).Result()
	if err != nil {
		return domain.Order{}, err
	}
	var order domain.Order
	err = json.Unmarshal([]byte(val), &order)
	return order, err
}

func (c *RedisOrderCache) Invalidate(ctx context.Context, id string) error {
	return c.client.Del(ctx, "order:"+id).Err()
}
