//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/KolManis/library-service/internal/adapters/out/repositories"
	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
)

type LoanRepositorySuite struct {
	IntegrationSuite
	repo *repositories.LoanRepository
}

func (s *LoanRepositorySuite) SetupTest() {
	s.IntegrationSuite.SetupTest() // TRUNCATE из базового сьюта
	s.repo = repositories.NewLoanRepository(s.db)
}

// TestCreateAndGet — полный круг через настоящую БД:
// домен → toModel → INSERT → SELECT → Restore → домен.
// Ловит рассинхрон миграции, модели и маппинга разом.
func (s *LoanRepositorySuite) TestCreateAndGet() {
	ctx := context.Background()
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	copyID, readerID := uuid.New(), uuid.New()
	s.insertLoanFixtures(copyID, readerID)

	l, err := loan.Reserve(copyID, readerID, now)
	s.Require().NoError(err)
	s.Require().NoError(s.repo.Create(ctx, l))

	got, err := s.repo.GetByID(ctx, l.ID())
	s.Require().NoError(err)

	s.Require().Equal(l.ID(), got.ID())
	s.Require().Equal(copyID, got.CopyID())
	s.Require().Equal(readerID, got.ReaderID())
	s.Require().Equal(loan.StatusReserved, got.Status())
	// время не сравниваем через Equal: Postgres хранит микросекунды,
	// Go — наносекунды, плюс таймзона драйвера
	s.Require().WithinDuration(now, got.ReservedAt(), time.Second)
	s.Require().Nil(got.DueAt(), "у брони ещё нет срока возврата")
}

// TestUpdate — переход статуса доезжает до БД и возвращается.
func (s *LoanRepositorySuite) TestUpdate() {
	ctx := context.Background()
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	copyID, readerID := uuid.New(), uuid.New()
	s.insertLoanFixtures(copyID, readerID)

	l, err := loan.Reserve(copyID, readerID, now)
	s.Require().NoError(err)
	s.Require().NoError(s.repo.Create(ctx, l))

	s.Require().NoError(l.Issue(now))
	s.Require().NoError(s.repo.Update(ctx, l))

	got, err := s.repo.GetByID(ctx, l.ID())
	s.Require().NoError(err)
	s.Require().Equal(loan.StatusIssued, got.Status())
	s.Require().NotNil(got.DueAt())
	s.Require().WithinDuration(now.AddDate(0, 0, loan.LoanPeriodDays), *got.DueAt(), time.Second)
}

// TestGetByID_NotFound — контракт порта: именно ErrLoanNotFound, не голая
// ошибка gorm. На этой ошибке команды будут строить ответ 404.
func (s *LoanRepositorySuite) TestGetByID_NotFound() {
	_, err := s.repo.GetByID(context.Background(), uuid.New())
	s.Require().ErrorIs(err, ports.ErrLoanNotFound)
}

// Мост в стандартный go test.
func TestLoanRepositorySuite(t *testing.T) {
	suite.Run(t, new(LoanRepositorySuite))
}
