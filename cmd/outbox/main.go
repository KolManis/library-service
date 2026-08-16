package main

import (
	"log/slog"
	"time"

	"github.com/KolManis/library-service/internal/adapters/out/producers"
	"github.com/KolManis/library-service/internal/adapters/out/repositories"
	"github.com/KolManis/library-service/internal/config"
	"github.com/KolManis/library-service/internal/infra"
)

func main() {
	infra.NewLogger()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("ошибка конфигурации", "error", err)
		return
	}

	db, err := infra.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		slog.Error("подключение к БД", "error", err)
		return
	}

	producer, err := infra.NewKafkaSyncProducer(cfg.KafkaBrokers)
	if err != nil {
		slog.Error("подключение к Kafka", "error", err)
		return
	}

	defer func() {
		if err := producer.Close(); err != nil {
			slog.Error("закрытие Kafka-продюсера", "error", err)
		}
	}()

	kafkaPublisher := producers.NewKafkaPublisher(producer)

	outboxRepo := repositories.NewOutboxRepository(db)
	ctx, stop := infra.ShutdownContext()
	defer stop()

	interval := 5 * time.Second
	slog.Info("outbox worker запущен", "interval", interval)
	producers.Run(ctx, outboxRepo, kafkaPublisher, interval)

	slog.Info("outbox worker остановлен")
}
