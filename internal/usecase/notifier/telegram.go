package notifier

import (
	"context"
	"strconv"

	"delayed-notifier/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wb-go/wbf/zlog"
)

type TelegramNotifier struct {
	cfg TelegramConfig
}

type TelegramConfig struct {
	BotToken string
}

func NewTelegramNotifier(cfg TelegramConfig) *TelegramNotifier {
	return &TelegramNotifier{cfg: cfg}
}

func (t *TelegramNotifier) Send(ctx context.Context, notification *domain.Notification) error {
	bot, err := tgbotapi.NewBotAPI(t.cfg.BotToken)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("id", notification.ID).Msg("Failed to create Telegram bot")
		return err
	}
	chatID, err := strconv.ParseInt(notification.UserID, 10, 64)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("id", notification.ID).Msg("Invalid chat ID")
		return err
	}
	msg := tgbotapi.NewMessage(chatID, notification.Message)
	zlog.Logger.Info().
		Int64("chat_id", chatID).
		Str("channel", "telegram").
		Str("id", notification.ID).
		Msg("Sending Telegram notification")
	_, err = bot.Send(msg)
	return err
}
