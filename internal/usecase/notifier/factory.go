package notifier

import (
	"context"
	"fmt"

	"delayed-notifier/internal/config"
	"delayed-notifier/internal/domain"
)

type MultiNotifier struct {
	Email    *EmailNotifier
	Telegram *TelegramNotifier
}

func NewMultiNotifier(cfg *config.Config) *MultiNotifier {
	return &MultiNotifier{
		Email: NewEmailNotifier(EmailConfig{
			SmtpHost: cfg.Email.SmtpHost,
			SmtpPort: cfg.Email.SmtpPort,
			User:     cfg.Email.User,
			Pass:     cfg.Email.Pass,
		}),
		Telegram: NewTelegramNotifier(TelegramConfig{
			BotToken: cfg.Telegram.BotToken,
		}),
	}
}

func (m *MultiNotifier) Send(ctx context.Context, notification *domain.Notification) error {
	switch notification.Channel {
	case domain.ChannelEmail:
		if m.Email == nil {
			return fmt.Errorf("email notifier not configured")
		}
		return m.Email.Send(ctx, notification)
	case domain.ChannelTelegram:
		if m.Telegram == nil {
			return fmt.Errorf("telegram notifier not configured")
		}
		return m.Telegram.Send(ctx, notification)
	default:
		return domain.ErrUnknownChannel
	}
}
