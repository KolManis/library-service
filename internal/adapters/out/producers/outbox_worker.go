package producers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/fine"
	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
	eventsv1 "github.com/KolManis/library-service/pkg/events/v1"
)

// publishRecord обрабатывает одну запись outbox: JSON -> protobuf -> Kafka.
func publishRecord(ctx context.Context, pub ports.IEventPublisher, rec ports.OutboxRecord) error {
	switch rec.EventType {
	case "loan.changed":
		var e loan.LoanChanged
		if err := json.Unmarshal(rec.Payload, &e); err != nil {
			return fmt.Errorf("распаковка loan.changed: %w", err)
		}
		msg := &eventsv1.LoanChanged{
			LoanId:     e.LoanID.String(),
			Status:     string(e.Status),
			OccurredAt: timestamppb.New(e.OccurredAt),
		}
		return publishProto(ctx, pub, "loan.changed", e.LoanID.String(), msg)

	case "fine.created":
		var e fine.FineCreated
		if err := json.Unmarshal(rec.Payload, &e); err != nil {
			return fmt.Errorf("распаковка fine.created: %w", err)
		}
		msg := &eventsv1.FineCreated{
			FineId:     e.FineID.String(),
			LoanId:     e.LoanID.String(),
			Amount:     fmt.Sprintf("%.2f", e.Amount),
			OccurredAt: timestamppb.New(e.OccurredAt),
		}
		return publishProto(ctx, pub, "fine.created", e.FineID.String(), msg)

	default:
		return fmt.Errorf("неизвестный event_type: %s", rec.EventType)
	}
}

// publishProto сериализует protobuf и отправляет в Kafka.
func publishProto(ctx context.Context, pub ports.IEventPublisher, topic, key string, msg proto.Message) error {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("маршалинг protobuf: %w", err)
	}
	return pub.Publish(ctx, topic, key, payload)
}

// Run — цикл поллинга outbox. Блокируется до отмены контекста.
func Run(ctx context.Context, reader ports.IOutboxReader, pub ports.IEventPublisher, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("outbox worker остановлен")
			return
		case <-ticker.C:
			Tick(ctx, reader, pub)
		}
	}
}

// Tick забирает неопубликованные записи, публикует и отмечает их.
func Tick(ctx context.Context, reader ports.IOutboxReader, pub ports.IEventPublisher) {
	records, err := reader.FetchUnpublished(ctx, 100)
	if err != nil {
		slog.Error("чтение outbox", "error", err)
		return
	}

	for _, rec := range records {
		if err := publishRecord(ctx, pub, rec); err != nil {
			slog.Error("публикация события", "id", rec.ID, "error", err)
			continue // не отмечаем published_at, повторим на следующем тике
		}
		if err := reader.MarkPublished(ctx, rec.ID); err != nil {
			slog.Error("отметка published_at", "id", rec.ID, "error", err)
		}
	}
}
