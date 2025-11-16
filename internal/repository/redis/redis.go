package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"delayed-notifier/internal/config"
	"delayed-notifier/internal/domain"

	wbfredis "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type RedisCache struct {
	client  *wbfredis.Client
	retries retry.Strategy
}

func NewRedisCache(cfg *config.Config, retries retry.Strategy) *RedisCache {
	client := wbfredis.New(cfg.RedisAddr(), cfg.Redis.Pass, cfg.Redis.DB)
	return &RedisCache{
		client:  client,
		retries: retries,
	}
}

func (r *RedisCache) Get(ctx context.Context, id string) (*domain.Notification, error) {
	val, err := r.client.GetWithRetry(ctx, r.retries, "notif:"+id)
	if err != nil {
		return nil, fmt.Errorf("failed to get from redis: %w", err)
	}
	if val == "" {
		return nil, nil
	}
	var notif domain.Notification
	if err := json.Unmarshal([]byte(val), &notif); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification: %w", err)
	}
	return &notif, nil
}

func (r *RedisCache) Set(ctx context.Context, id string, notif *domain.Notification, ttl time.Duration) error {
	data, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}
	if err := r.client.SetWithRetry(ctx, r.retries, "notif:"+id, string(data)); err != nil {
		return fmt.Errorf("failed to set in redis: %w", err)
	}
	return nil
}

func (r *RedisCache) Del(ctx context.Context, id string) error {
	if err := r.client.DelWithRetry(ctx, r.retries, "notif:"+id); err != nil {
		return fmt.Errorf("failed to delete from redis: %w", err)
	}
	return nil
}

func (r *RedisCache) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close redis client: %w", err)
	}
	return nil
}
