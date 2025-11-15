package rabbitmq

import (
	"delayed-notifier/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
	wbfrabbit "github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type Broker struct {
	client  *wbfrabbit.RabbitClient
	retries retry.Strategy
}

func NewRabbitMQ(cfg *config.Config, retries retry.Strategy) *Broker {
	rabbitCfg := wbfrabbit.ClientConfig{
		URL:            cfg.RabbitMQDSN(),
		ConnectTimeout: cfg.RabbitMQ.ConnectTimeout,
		Heartbeat:      cfg.RabbitMQ.Heartbeat,
		PublishRetry:   retries,
		ConsumeRetry:   retries,
	}
	client, err := wbfrabbit.NewClient(rabbitCfg)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to create RabbitMQ client")
	}
	ch, err := client.GetChannel()
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to get channel for declarations")
	}
	defer ch.Close()
	err = ch.ExchangeDeclare("delayed_notifications", "x-delayed-message", true, false, false, false, amqp.Table{"x-delayed-type": "direct"})
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to declare exchange")
	}
	_, err = ch.QueueDeclare("notifications", true, false, false, false, nil)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to declare queue")
	}
	err = ch.QueueBind("notifications", "notify", "delayed_notifications", false, nil)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to bind queue")
	}
	_, err = ch.QueueDeclare("notifications_dlq", true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "delayed_notifications",
		"x-dead-letter-routing-key": "notify",
		"x-message-ttl":             1000,
	})
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to declare DLQ")
	}
	return &Broker{client: client, retries: retries}
}

func (b *Broker) Close() error {
	zlog.Logger.Info().Msg("Closing RabbitMQ connection")
	return b.client.Close()
}
