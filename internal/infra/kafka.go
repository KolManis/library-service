package infra

import (
	"fmt"

	"github.com/IBM/sarama"
)

// NewKafkaSyncProducer создаёт синхронного продюсера Kafka.
func NewKafkaSyncProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("подключение к Kafka: %w", err)
	}
	return producer, nil
}
