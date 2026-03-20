package kafkax

import (
	"context"

	kafkago "github.com/segmentio/kafka-go"
)

type MessageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafkago.Message) error
	Close() error
}

func NewWriter(brokers []string, topic string) *kafkago.Writer {
	return &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequireOne,
		Async:        false,
	}
}
