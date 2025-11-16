// internal/broker/rabbitmq/rabbitmq.go
package rabbitmq

import (
	"context"
	"delayed-notifier/internal/config"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	wbfrabbit "github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type RabbitMQ struct {
	client    *wbfrabbit.RabbitClient
	retries   retry.Strategy
	consumer  *Consumer
	publisher *Publisher
}

func NewRabbitMQ(cfg *config.Config, retries retry.Strategy) (*RabbitMQ, error) {
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

	err = ch.QueueBind("notifications", "notify", "delayed_notifications", false, nil)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to bind queue")
	}

	_, err = ch.QueueDeclare("notifications_dlq", true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "delayed_notifications",
		"x-dead-letter-routing-key": "notify",
		"x-message-ttl":             60000,
	})
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to declare DLQ")
	}
	return &RabbitMQ{client: client, retries: retries}, nil
}

func (b *RabbitMQ) Publish(ctx context.Context, exchange, key string, body []byte) error {
	if b.publisher == nil {
		b.publisher = NewPublisher(b.client)
	}
	if exchange != "delayed_notifications" {
		return errors.New("unsupported exchange")
	}
	return b.publisher.publisher.Publish(ctx, body, key)
}

func (b *RabbitMQ) PublishDelayed(ctx context.Context, id string, delay time.Duration) error {
	if b.publisher == nil {
		b.publisher = NewPublisher(b.client)
	}
	return b.publisher.PublishDelayed(ctx, id, delay)
}

func (b *RabbitMQ) Consume(ctx context.Context, queue string, handler wbfrabbit.MessageHandler) error {
	cfg := wbfrabbit.ConsumerConfig{
		Queue:         queue,
		ConsumerTag:   "notification_consumer",
		AutoAck:       false,
		Workers:       5,
		PrefetchCount: 10,
		Nack:          wbfrabbit.NackConfig{Multiple: false, Requeue: true},
		Ask:           wbfrabbit.AskConfig{Multiple: false},
	}
	b.consumer = NewConsumer(b.client, cfg, handler)
	return b.consumer.Consume(ctx)
}

func (b *RabbitMQ) Close() error {
	zlog.Logger.Info().Msg("Closing RabbitMQ connection")
	return b.client.Close()
}
