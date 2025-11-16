package repository

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

type NotificationRepository interface {
	Create(ctx context.Context, notif *domain.Notification) error
	Get(ctx context.Context, id string) (*domain.Notification, error)
	UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus) error
	IncrementRetry(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*domain.Notification, error)
	GetPendingNotifications(ctx context.Context) ([]*domain.Notification, error)
}
