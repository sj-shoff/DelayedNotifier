package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"delayed-notifier/internal/domain"
	"delayed-notifier/internal/repository"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type NotificationRepository struct {
	db      *dbpg.DB
	cache   repository.Cache
	retries retry.Strategy
	ttl     time.Duration
}

func NewNotificationRepository(
	db *dbpg.DB,
	cache repository.Cache,
	retries retry.Strategy,
	ttl time.Duration,
) *NotificationRepository {
	r := &NotificationRepository{
		db:      db,
		cache:   cache,
		retries: retries,
		ttl:     ttl,
	}

	return r
}

func (r *NotificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	_, err := r.db.ExecWithRetry(ctx, r.retries,
		`INSERT INTO notifications (id, user_id, channel, message, send_at, status, retries, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		notif.ID, notif.UserID, notif.Channel, notif.Message, notif.SendAt,
		notif.Status, notif.Retries, notif.CreatedAt, notif.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	r.cache.Set(ctx, notif.ID, notif, r.ttl)
	return nil
}

func (r *NotificationRepository) Get(ctx context.Context, id string) (*domain.Notification, error) {
	cached, err := r.cache.Get(ctx, id)
	if err == nil && cached != nil {
		return cached, nil
	}
	row, err := r.db.QueryRowWithRetry(ctx, r.retries,
		`SELECT id, user_id, channel, message, send_at, status, retries, created_at, updated_at
FROM notifications WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query notification: %w", err)
	}
	var notif domain.Notification
	err = row.Scan(
		&notif.ID, &notif.UserID, &notif.Channel, &notif.Message, &notif.SendAt,
		&notif.Status, &notif.Retries, &notif.CreatedAt, &notif.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan notification: %w", err)
	}
	r.cache.Set(ctx, id, &notif, r.ttl)
	return &notif, nil
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus) error {
	_, err := r.db.ExecWithRetry(ctx, r.retries,
		`UPDATE notifications SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	r.cache.Del(ctx, id)
	return nil
}

func (r *NotificationRepository) IncrementRetry(ctx context.Context, id string) error {
	_, err := r.db.ExecWithRetry(ctx, r.retries,
		`UPDATE notifications SET retries = retries + 1, updated_at = $1 WHERE id = $2`,
		time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to increment retry: %w", err)
	}
	r.cache.Del(ctx, id)
	return nil
}

func (r *NotificationRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecWithRetry(ctx, r.retries,
		`DELETE FROM notifications WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	r.cache.Del(ctx, id)
	return nil
}

func (r *NotificationRepository) List(ctx context.Context) ([]*domain.Notification, error) {
	rows, err := r.db.QueryWithRetry(ctx, r.retries,
		`SELECT id, user_id, channel, message, send_at, status, retries, created_at, updated_at
FROM notifications ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()
	var notifs []*domain.Notification
	for rows.Next() {
		var notif domain.Notification
		err := rows.Scan(
			&notif.ID, &notif.UserID, &notif.Channel, &notif.Message, &notif.SendAt,
			&notif.Status, &notif.Retries, &notif.CreatedAt, &notif.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification in list: %w", err)
		}
		notifs = append(notifs, &notif)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows in list: %w", err)
	}
	return notifs, nil
}

func (r *NotificationRepository) GetPendingNotifications(ctx context.Context) ([]*domain.Notification, error) {
	rows, err := r.db.QueryWithRetry(ctx, r.retries,
		`SELECT id, user_id, channel, message, send_at, status, retries, created_at, updated_at
			FROM notifications
			WHERE status = $1 AND send_at <= $2
			ORDER BY send_at ASC
			LIMIT 100`,
		domain.StatusPending, time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}
	defer rows.Close()
	var notifs []*domain.Notification
	for rows.Next() {
		var notif domain.Notification
		err := rows.Scan(
			&notif.ID, &notif.UserID, &notif.Channel, &notif.Message, &notif.SendAt,
			&notif.Status, &notif.Retries, &notif.CreatedAt, &notif.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending notification: %w", err)
		}
		notifs = append(notifs, &notif)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows in pending: %w", err)
	}
	return notifs, nil
}
