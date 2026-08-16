//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/KolManis/library-service/internal/adapters/out/producers"
	"github.com/KolManis/library-service/internal/adapters/out/repositories"
	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	eventsv1 "github.com/KolManis/library-service/pkg/events/v1"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	"google.golang.org/protobuf/proto"
)

type OutboxWorkerSuite struct {
	IntegrationSuite
	brokers []string
}

func TestOutboxWorkerSuite(t *testing.T) {
	suite.Run(t, new(OutboxWorkerSuite))
}

func (s *OutboxWorkerSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()
	ctx := context.Background()
	container, err := tckafka.Run(ctx, "confluentinc/confluent-local:7.5.0")
	s.Require().NoError(err)
	s.brokers, err = container.Brokers(ctx)
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			s.T().Logf("завершение Kafka-контейнера: %v", err)
		}
	})
}

func (s *OutboxWorkerSuite) TestOutboxWorker_PublishesToKafka() {
	// 1. Домен создаёт Loan, событие вручную кладём в outbox — так же,
	//    как это делает reservecopy.Handler внутри транзакции.
	l, err := loan.Reserve(fixtureCopyID, fixtureReaderID, time.Now())
	s.Require().NoError(err)
	s.Require().NoError(repositories.NewLoanRepository(s.db).Create(context.Background(), l))
	outboxRepo := repositories.NewOutboxRepository(s.db)
	s.Require().NoError(outboxRepo.Append(context.Background(), l.PullEvents()))

	// 2. Настоящий sarama-продюсер на testcontainers-брокер
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(s.brokers, config)
	s.Require().NoError(err)
	defer func() {
		if err := producer.Close(); err != nil {
			s.T().Logf("закрытие producer: %v", err)
		}
	}()

	// 3. Один тик воркера
	producers.Tick(context.Background(), outboxRepo, producers.NewKafkaPublisher(producer))

	// 4. published_at проставился
	var count int64
	s.db.Raw(`SELECT COUNT(*) FROM outbox WHERE published_at IS NOT NULL`).Scan(&count)
	s.Require().Equal(int64(1), count)

	// 5. Сообщение реально долетело до топика — читаем настоящим консьюмером
	consumer, err := sarama.NewConsumer(s.brokers, sarama.NewConfig())
	s.Require().NoError(err)
	defer func() {
		if err := consumer.Close(); err != nil {
			s.T().Logf("закрытие consumer: %v", err)
		}
	}()

	partConsumer, err := consumer.ConsumePartition("loan.changed", 0, sarama.OffsetOldest)
	s.Require().NoError(err)
	defer func() {
		if err := partConsumer.Close(); err != nil {
			s.T().Logf("закрытие partConsumer: %v", err)
		}
	}()

	select {
	case msg := <-partConsumer.Messages():
		s.Equal(l.ID().String(), string(msg.Key)) // ключ = id агрегата
		var got eventsv1.LoanChanged
		s.Require().NoError(proto.Unmarshal(msg.Value, &got))
		s.Equal(l.ID().String(), got.LoanId)
		s.Equal(string(loan.StatusReserved), got.Status)
	case <-time.After(10 * time.Second):
		s.Fail("сообщение не пришло в топик за 10 секунд")
	}
}
