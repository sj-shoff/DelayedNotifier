package broker

import (
	"context"
	"time"

	wbfrabbit "github.com/wb-go/wbf/rabbitmq"
)

type Broker interface {
	Publish(ctx context.Context, exchange, key string, body []byte) error
	PublishDelayed(ctx context.Context, notificationID string, delay time.Duration) error
	Consume(ctx context.Context, queue string, handler wbfrabbit.MessageHandler) error
	Close() error
}
