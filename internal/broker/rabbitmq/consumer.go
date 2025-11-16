// internal/broker/rabbitmq/consumer.go
package rabbitmq

import (
	"context"

	wbfrabbit "github.com/wb-go/wbf/rabbitmq"
)

type Consumer struct {
	consumer *wbfrabbit.Consumer
}

func NewConsumer(client *wbfrabbit.RabbitClient, cfg wbfrabbit.ConsumerConfig, handler wbfrabbit.MessageHandler) *Consumer {
	return &Consumer{
		consumer: wbfrabbit.NewConsumer(client, cfg, handler),
	}
}

func (c *Consumer) Consume(ctx context.Context) error {
	return c.consumer.Start(ctx)
}
