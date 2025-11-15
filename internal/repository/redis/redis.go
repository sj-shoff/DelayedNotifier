package redis

import (
	"context"
	"encoding/json"
	"time"

	"delayed-notifier/internal/domain"

	wbfredis "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type RedisCache struct {
	client  *wbfredis.Client
	retries retry.Strategy
}

func NewRedisCache(client *wbfredis.Client, retries retry.Strategy) *RedisCache {
	return &RedisCache{
		client:  client,
		retries: retries,
	}
}

func (r *RedisCache) Get(ctx context.Context, id string) (*domain.Notification, error) {
	val, err := r.client.GetWithRetry(ctx, r.retries, "notif:"+id)
	if err != nil || val == "" {
		return nil, err
	}
	var notif domain.Notification
	if err := json.Unmarshal([]byte(val), &notif); err != nil {
		return nil, err
	}
	return &notif, nil
}

func (r *RedisCache) Set(ctx context.Context, id string, notif *domain.Notification, ttl time.Duration) error {
	data, err := json.Marshal(notif)
	if err != nil {
		return err
	}
	return r.client.SetWithRetry(ctx, r.retries, "notif:"+id, string(data))
}

func (r *RedisCache) Del(ctx context.Context, id string) error {
	return r.client.DelWithRetry(ctx, r.retries, "notif:"+id)
}
