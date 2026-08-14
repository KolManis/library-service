//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/KolManis/library-service/internal/adapters/out/repositories"
	"github.com/KolManis/library-service/internal/core/application/commands/reservecopy"
	"github.com/KolManis/library-service/internal/core/domain/events"
)

// failingOutboxRepo — специально ломает Append, чтобы проверить откат транзакции.
type failingOutboxRepo struct{}

func (failingOutboxRepo) Append(ctx context.Context, evs []events.DomainEvent) error {
	return errors.New("симулированный сбой записи в outbox")
}

type OutboxAtomicitySuite struct {
	IntegrationSuite
}

func (s *OutboxAtomicitySuite) TestReserve_OutboxFails_RollsBackLoan() {
	copyRepo := repositories.NewCopyRepository(s.db)
	loanRepo := repositories.NewLoanRepository(s.db)
	transactor := repositories.NewTransactor(s.db)
	handler := reservecopy.NewHandler(copyRepo, loanRepo, failingOutboxRepo{}, transactor, time.Now)

	cmd, _ := reservecopy.NewCommand(fixtureBookID, fixtureReaderID)
	_, err := handler.Handle(context.Background(), cmd)
	s.Require().Error(err)

	// Прямым SQL, в обход репозиториев — ни Loan, ни outbox не должны были появиться
	var loanCount int64
	s.db.Raw(`SELECT COUNT(*) FROM loans`).Scan(&loanCount)
	s.Require().Equal(int64(0), loanCount, "Create должен был откатиться вместе с Append")

	var outboxCount int64
	s.db.Raw(`SELECT COUNT(*) FROM outbox`).Scan(&outboxCount)
	s.Require().Equal(int64(0), outboxCount)
}

// Мост в стандартный go test.
func TestOutboxAtomicitySuite(t *testing.T) {
	suite.Run(t, new(OutboxAtomicitySuite))
}
