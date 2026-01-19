package cache

import (
	"context"
	"delayed-notifier/internal/domain"
	"time"
)

type Cache interface {
	Set(ctx context.Context, id string, notif *domain.Notification, ttl time.Duration) error
	Del(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*domain.Notification, error)
	Close() error
}
