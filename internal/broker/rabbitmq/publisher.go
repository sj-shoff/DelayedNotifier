package rabbitmq

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

func (b *Broker) Publish(ctx context.Context, exchange, key string, publishing amqp.Publishing) error {
	return retry.DoContext(ctx, b.retries, func() error {
		ch, err := b.client.GetChannel()
		if err != nil {
			return err
		}
		defer ch.Close()
		return ch.PublishWithContext(ctx, exchange, key, false, false, publishing)
	})
}

func (b *Broker) PublishDelayed(ctx context.Context, notificationID string, delay time.Duration) error {
	publishing := amqp.Publishing{
		Body: []byte(notificationID),
		Headers: amqp.Table{
			"x-delay": int32(delay.Milliseconds()),
		},
	}
	return b.Publish(ctx, "delayed_notifications", "notify", publishing)
}
