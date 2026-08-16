package ports

import (
	"context"
)

// IEventPublisher — порт для публикации сериализованных событий в брокер.
// Не знает про protobuf/Kafka — просто топик, ключ (для партиционирования
// по агрегату) и байты.
type IEventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload []byte) error
}
