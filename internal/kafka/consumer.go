package kafka

import "github.com/segmentio/kafka-go"

type Consumer struct {
	Reader *kafka.Reader
}

func NewConsumer(brokers []string, topic string) *Consumer {
	return &Consumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic: topic,
			GroupID: "webhook-worker",
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
	}
}

func (c *Consumer) Close() error {
	return c.Reader.Close()
}