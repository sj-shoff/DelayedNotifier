package usecase

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
