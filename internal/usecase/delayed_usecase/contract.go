package delayed_usecase

import (
	"context"
	"time"

	"delayed-notifier/internal/domain"
)

type MessageBroker interface {
	PublishDelayed(ctx context.Context, id string, delay time.Duration) error
}

type Notifier interface {
	Send(ctx context.Context, notification *domain.Notification) error
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
