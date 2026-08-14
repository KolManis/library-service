//go:build integration

package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/KolManis/library-service/internal/adapters/out/repositories"
	"github.com/KolManis/library-service/internal/core/application/commands/reservecopy"
	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/stretchr/testify/suite"
)

type RaceSuite struct {
	IntegrationSuite
}

// SetupTest переопределяет базовый: после очистки и загрузки фикстур занимаем
// fixtureCopyID2 бронью, чтобы для гонки остался только fixtureCopyID.
func (s *RaceSuite) SetupTest() {
	s.IntegrationSuite.SetupTest() // TRUNCATE + загрузка фикстур

	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	l, err := loan.Reserve(fixtureCopyID2, fixtureReaderID, now)
	s.Require().NoError(err)
	repo := repositories.NewLoanRepository(s.db)
	s.Require().NoError(repo.Create(context.Background(), l))
}

// TestReserve_ConcurrentRequests_NoOverbooking проверяет, что при одновременных
// запросах на бронирование книги с одним свободным экземпляром успешной будет
// только ОДНА бронь. Без частичного уникального индекса (миграция 00002)
// этот тест должен падать с овербукингом (successCount > 1).
func (s *RaceSuite) TestReserve_ConcurrentRequests_NoOverbooking() {
	copyRepo := repositories.NewCopyRepository(s.db)
	loanRepo := repositories.NewLoanRepository(s.db)
	outboxRepo := repositories.NewOutboxRepository(s.db)
	transactor := repositories.NewTransactor(s.db)
	handler := reservecopy.NewHandler(copyRepo, loanRepo, outboxRepo, transactor, func() time.Time {
		return time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	})

	const n = 100
	var wg sync.WaitGroup
	var start sync.WaitGroup
	errs := make([]error, n)

	start.Add(1)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start.Wait()
			cmd, err := reservecopy.NewCommand(fixtureBookID, fixtureReaderID)
			if err != nil {
				errs[i] = err
				return
			}
			_, errs[i] = handler.Handle(context.Background(), cmd)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)
	start.Done()
	wg.Wait()

	successCount := 0
	for _, err := range errs {
		if err == nil {
			successCount++
		}
	}
	s.Require().Equal(1, successCount, "ровно один участник гонки должен получить экземпляр")
}

// TestRaceSuite запускает сьют гонок.
func TestRaceSuite(t *testing.T) {
	suite.Run(t, new(RaceSuite))
}
