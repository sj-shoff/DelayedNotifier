package handler

import (
	"context"
	"delayed-notifier/internal/domain"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, notification *domain.CreateNotification) (*domain.Notification, error)
	GetNotificationStatus(ctx context.Context, id string) (domain.NotificationStatus, error)
	CancelNotification(ctx context.Context, id string) error
	ListNotifications(ctx context.Context) ([]*domain.Notification, error)
	ProcessNotification(ctx context.Context, id string) error
}
