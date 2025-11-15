package dto

import (
	"time"

	"delayed-notifier/internal/domain"
)

type CreateNotificationRequest struct {
	UserID  string `json:"user_id" validate:"required"`
	Channel string `json:"channel" validate:"required,oneof=email telegram"`
	Message string `json:"message" validate:"required"`
	SendAt  string `json:"send_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

type NotificationResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Channel   string    `json:"channel"`
	Message   string    `json:"message"`
	SendAt    time.Time `json:"send_at"`
	Status    string    `json:"status"`
	Retries   int       `json:"retries"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StatusResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func FromDomain(n *domain.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		Channel:   string(n.Channel),
		Message:   n.Message,
		SendAt:    n.SendAt,
		Status:    string(n.Status),
		Retries:   n.Retries,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

func ToDomain(req CreateNotificationRequest) (*domain.CreateNotification, error) {
	sendAt, err := time.Parse(time.RFC3339, req.SendAt)
	if err != nil {
		return nil, err
	}
	if sendAt.Before(time.Now()) {
		return nil, domain.ErrSendAtInPast
	}
	return &domain.CreateNotification{
		UserID:  req.UserID,
		Channel: domain.NotificationChannel(req.Channel),
		Message: req.Message,
		SendAt:  sendAt,
	}, nil
}
