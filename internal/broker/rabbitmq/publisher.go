// internal/broker/rabbitmq/publisher.go
package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	wbfrabbit "github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

type Publisher struct {
	publisher *wbfrabbit.Publisher
}

func NewPublisher(client *wbfrabbit.RabbitClient) *Publisher {
	return &Publisher{
		publisher: wbfrabbit.NewPublisher(client, "delayed_notifications", "application/json"),
	}
}

func (p *Publisher) PublishDelayed(ctx context.Context, id string, delay time.Duration) error {
	payload := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to marshal payload")
		return err
	}

	delayMs := int(delay.Milliseconds())
	headers := amqp.Table{"x-delay": delayMs}

	zlog.Logger.Info().Str("id", id).Int("delay_ms", delayMs).Msg("Publishing delayed message")

	return p.publisher.Publish(ctx, body, "notify", wbfrabbit.WithHeaders(headers))
}
