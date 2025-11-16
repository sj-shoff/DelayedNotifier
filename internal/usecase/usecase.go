package usecase

import (
	"context"
	"math"
	"time"

	"delayed-notifier/internal/domain"
	"delayed-notifier/internal/repository"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type NotificationUsecase struct {
	repo     repository.NotificationRepository
	broker   MessageBroker
	retries  retry.Strategy
	notifier Notifier
}

func NewNotificationUsecase(
	repo repository.NotificationRepository,
	broker MessageBroker,
	retries retry.Strategy,
	notifier Notifier,
) *NotificationUsecase {
	return &NotificationUsecase{
		repo:     repo,
		broker:   broker,
		retries:  retries,
		notifier: notifier,
	}
}

func (u *NotificationUsecase) CreateNotification(ctx context.Context, dto *domain.CreateNotification) (*domain.Notification, error) {
	if dto.SendAt.Before(time.Now()) {
		return nil, domain.ErrSendAtInPast
	}
	notif := &domain.Notification{
		ID:        uuid.New().String(),
		UserID:    dto.UserID,
		Channel:   dto.Channel,
		Message:   dto.Message,
		SendAt:    dto.SendAt,
		Status:    domain.StatusPending,
		Retries:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := u.repo.Create(ctx, notif); err != nil {
		return nil, err
	}
	delay := time.Until(notif.SendAt)
	if err := u.broker.PublishDelayed(ctx, notif.ID, delay); err != nil {
		zlog.Logger.Warn().Err(err).Str("id", notif.ID).Msg("Failed to publish to broker")
	}
	return notif, nil
}

func (u *NotificationUsecase) GetNotificationStatus(ctx context.Context, id string) (domain.NotificationStatus, error) {
	notif, err := u.repo.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if notif == nil {
		return "", domain.ErrNotFound
	}
	return notif.Status, nil
}

func (u *NotificationUsecase) CancelNotification(ctx context.Context, id string) error {
	notif, err := u.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if notif == nil {
		return domain.ErrNotFound
	}
	if notif.Status != domain.StatusPending {
		return domain.ErrCannotCancel
	}
	return u.repo.UpdateStatus(ctx, id, domain.StatusCancelled)
}

func (u *NotificationUsecase) ListNotifications(ctx context.Context) ([]*domain.Notification, error) {
	return u.repo.List(ctx)
}

func (u *NotificationUsecase) ProcessNotification(ctx context.Context, id string) error {
	notif, err := u.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if notif == nil {
		return domain.ErrNotFound
	}
	if notif.Status != domain.StatusPending {
		zlog.Logger.Info().Str("id", id).Str("status", string(notif.Status)).Msg("Notification already processed")
		return nil
	}
	if notif.SendAt.After(time.Now()) {
		delay := time.Until(notif.SendAt)
		return u.broker.PublishDelayed(ctx, id, delay)
	}
	err = retry.DoContext(ctx, u.retries, func() error {
		return u.notifier.Send(ctx, notif)
	})
	if err != nil {
		zlog.Logger.Error().Err(err).Str("id", id).Msg("Failed to send notification")
		if err := u.repo.IncrementRetry(ctx, id); err != nil {
			return err
		}
		updatedNotif, _ := u.repo.Get(ctx, id)
		if updatedNotif.Retries >= u.retries.Attempts {
			return u.repo.UpdateStatus(ctx, id, domain.StatusFailed)
		}
		delay := u.retries.Delay * time.Duration(math.Pow(u.retries.Backoff, float64(updatedNotif.Retries-1)))
		return u.broker.PublishDelayed(ctx, id, delay)
	}
	return u.repo.UpdateStatus(ctx, id, domain.StatusSent)
}
