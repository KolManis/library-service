package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/KolManis/library-service/internal/core/domain/events"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OutboxRepository — gorm-реализация ports.IOutboxRepository.
type OutboxRepository struct {
	db *gorm.DB
}

// NewOutboxRepository создаёт OutboxRepository поверх открытого соединения gorm.
func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// outboxModel — строка таблицы outbox. Знает про БД всё,
// про бизнес-правила — ничего.
type outboxModel struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	EventType   string
	Payload     []byte
	CreatedAt   time.Time
	PublishedAt *time.Time
}

// TableName говорит gorm имя таблицы — иначе он выведет "outbox_models".
func (outboxModel) TableName() string { return "outbox" }

// Append сохраняет события. Каждое сериализуется в JSON и пишется отдельной
// строкой; вызывается внутри той же транзакции, что и сохранение агрегата.
func (r *OutboxRepository) Append(ctx context.Context, evs []events.DomainEvent) error {
	for _, ev := range evs {
		payload, err := json.Marshal(ev)
		if err != nil {
			return err
		}
		m := outboxModel{
			ID:        uuid.New(),
			EventType: ev.EventType(),
			Payload:   payload,
			CreatedAt: time.Now(),
		}
		if err := dbFromContext(ctx, r.db).Create(&m).Error; err != nil {
			return err
		}
	}
	return nil
}
