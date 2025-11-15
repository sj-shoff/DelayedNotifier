package rabbitmq

import (
	"context"
	"time"

	"github.com/wb-go/wbf/zlog"
)

type ProcessFunc func(ctx context.Context, messageID string) error

type Worker struct {
	broker      *Broker
	processFunc ProcessFunc
	done        chan struct{}
}

func NewWorker(broker *Broker, processFunc ProcessFunc) *Worker {
	return &Worker{
		broker:      broker,
		processFunc: processFunc,
		done:        make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	zlog.Logger.Info().Msg("Starting notification worker")
	for {
		select {
		case <-ctx.Done():
			zlog.Logger.Info().Msg("Worker context cancelled")
			return
		case <-w.done:
			zlog.Logger.Info().Msg("Worker stopped")
			return
		default:
			w.processMessages(ctx)
			time.Sleep(5 * time.Second)
		}
	}
}

func (w *Worker) Stop() {
	close(w.done)
}

func (w *Worker) processMessages(ctx context.Context) {
	deliveries, err := w.broker.Consume(ctx, "notifications")
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to consume messages")
		return
	}
	for delivery := range deliveries {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		default:
			notificationID := string(delivery.Body)
			zlog.Logger.Info().Str("id", notificationID).Msg("Processing notification")
			if err := w.processFunc(ctx, notificationID); err != nil {
				zlog.Logger.Error().Err(err).Str("id", notificationID).Msg("Failed to process notification")
				delivery.Nack(false, false)
			} else {
				delivery.Ack(false)
				zlog.Logger.Info().Str("id", notificationID).Msg("Notification processed successfully")
			}
		}
	}
}
