package kafka

import (
	"errors"
	"fmt"

	"context"

	kafkago "github.com/segmentio/kafka-go"
)

func EnsureTopic(ctx context.Context, brokers []string, topic string, partitions int, replicationFactor int) error {
	if len(brokers) == 0 {
		return errors.New("no kafka brokers configured")
	}
	if topic == "" {
		return errors.New("kafka topic is empty")
	}
	if partitions <= 0 {
		partitions = 1
	}
	if replicationFactor <= 0 {
		replicationFactor = 1
	}

	var lastErr error
	for _, brokerAddr := range brokers {
		conn, err := kafkago.DialContext(ctx, "tcp", brokerAddr)
		if err != nil {
			lastErr = fmt.Errorf("dial broker %s: %w", brokerAddr, err)
			continue
		}

		controller, err := conn.Controller()
		_ = conn.Close()
		if err != nil {
			lastErr = fmt.Errorf("get controller from broker %s: %w", brokerAddr, err)
			continue
		}

		controllerConn, err := kafkago.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
		if err != nil {
			lastErr = fmt.Errorf("dial controller %s:%d: %w", controller.Host, controller.Port, err)
			continue
		}

		err = controllerConn.CreateTopics(kafkago.TopicConfig{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		})
		closeErr := controllerConn.Close()
		if err == nil || errors.Is(err, kafkago.TopicAlreadyExists) {
			return closeErr
		}
		if closeErr != nil {
			lastErr = errors.Join(err, closeErr)
			continue
		}

		lastErr = err
	}

	if lastErr == nil {
		lastErr = errors.New("failed to ensure kafka topic")
	}

	return lastErr
}
