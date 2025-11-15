package domain

import (
	"errors"
	"time"
)

type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusCancelled NotificationStatus = "cancelled"
	StatusFailed    NotificationStatus = "failed"
)

type NotificationChannel string

const (
	ChannelEmail    NotificationChannel = "email"
	ChannelTelegram NotificationChannel = "telegram"
)

type Notification struct {
	ID        string
	UserID    string
	Channel   NotificationChannel
	Message   string
	SendAt    time.Time
	Status    NotificationStatus
	Retries   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateNotification struct {
	UserID  string
	Channel NotificationChannel
	Message string
	SendAt  time.Time
}

var (
	ErrSendAtInPast   = errors.New("send_at must be in the future")
	ErrNotFound       = errors.New("notification not found")
	ErrCannotCancel   = errors.New("cannot cancel non-pending notification")
	ErrUnknownChannel = errors.New("unknown notification channel")
)
