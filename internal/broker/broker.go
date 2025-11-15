package broker

import "context"

type Broker interface {
	Publish(ctx context.Context, queue string, message []byte) error
	Consume(queue string, handler MessageHandler) error
	Close() error
}

type Worker interface {
	Start(ctx context.Context) error
	Stop() error
}

type MessageHandler func(ctx context.Context, message []byte) error
