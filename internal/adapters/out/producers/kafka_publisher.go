package producers

import (
	"context"

	"github.com/IBM/sarama"
)

// KafkaPublisher реализует ports.IEventPublisher поверх sarama.SyncProducer.
type KafkaPublisher struct {
	producer sarama.SyncProducer
}

// NewKafkaPublisher создаёт нового продюсера.
func NewKafkaPublisher(producer sarama.SyncProducer) *KafkaPublisher {
	return &KafkaPublisher{producer: producer}
}

// Publish отправляет сообщение в Kafka.
// key — ключ сообщения (id агрегата), payload — protobuf-байты.
func (p *KafkaPublisher) Publish(ctx context.Context, topic, key string, payload []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(payload),
	}
	_, _, err := p.producer.SendMessage(msg)
	return err
}
