package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

func (b *Broker) Consume(ctx context.Context, queue string) (<-chan amqp.Delivery, error) {
	var deliveries <-chan amqp.Delivery
	err := retry.DoContext(ctx, b.retries, func() error {
		ch, err := b.client.GetChannel()
		if err != nil {
			return err
		}
		defer ch.Close()
		deliveries, err = ch.Consume(queue, "", false, false, false, false, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return deliveries, nil
}
