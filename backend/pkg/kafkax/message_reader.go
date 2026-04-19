package kafkax

import "time"

import kafkago "github.com/segmentio/kafka-go"

func NewMessageReader(brokers []string, topic string, groupId string) *kafkago.Reader {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:                brokers,
		Topic:                  topic,
		GroupID:                groupId,
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Second,
	})
}
