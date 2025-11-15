package notifier

import (
	"context"
	"fmt"
	"net/smtp"

	"delayed-notifier/internal/domain"

	"github.com/wb-go/wbf/zlog"
)

type EmailNotifier struct {
	cfg EmailConfig
}

type EmailConfig struct {
	SmtpHost string
	SmtpPort int
	User     string
	Pass     string
}

func NewEmailNotifier(cfg EmailConfig) *EmailNotifier {
	return &EmailNotifier{cfg: cfg}
}

func (e *EmailNotifier) Send(ctx context.Context, notification *domain.Notification) error {
	auth := smtp.PlainAuth("", e.cfg.User, e.cfg.Pass, e.cfg.SmtpHost)
	to := []string{notification.UserID}
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: Notification\r\n"+
		"\r\n"+
		"%s\r\n", e.cfg.User, notification.UserID, notification.Message))
	addr := fmt.Sprintf("%s:%d", e.cfg.SmtpHost, e.cfg.SmtpPort)
	zlog.Logger.Info().
		Str("to", notification.UserID).
		Str("channel", "email").
		Str("id", notification.ID).
		Msg("Sending email notification")
	return smtp.SendMail(addr, auth, e.cfg.User, to, msg)
}
