package kafkax

import (
	"encoding/json"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Message = kafkago.Message

type WritableMessage interface {
	GetKey() string
	GetCreatedAt() time.Time
	GetMessageType() string
}

func BuildMessages[T WritableMessage](
	messages []T,
) ([]kafkago.Message, error) {
	kafkaMessages := make([]kafkago.Message, 0, len(messages))

	for _, message := range messages {
		payload, err := json.Marshal(message)
		if err != nil {
			return nil, fmt.Errorf("marshal message %s: %w", message.GetKey(), err)
		}

		kafkaMessages = append(kafkaMessages, kafkago.Message{
			Key:   []byte(message.GetKey()),
			Value: payload,
			Time:  message.GetCreatedAt(),
			Headers: []kafkago.Header{
				{Key: "message_type", Value: []byte(message.GetMessageType())},
			},
		})
	}

	return kafkaMessages, nil
}
