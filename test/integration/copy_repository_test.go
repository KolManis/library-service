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

// id из fixtures/*.yml (fixtureCopyID и fixtureReaderID объявлены
// в loan_repository_test.go — пакет общий)
var (
	fixtureBookID  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	fixtureCopyID2 = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

type CopyRepositorySuite struct {
	IntegrationSuite
	copies *repositories.CopyRepository
	loans  *repositories.LoanRepository
}

func (s *CopyRepositorySuite) SetupTest() {
	s.IntegrationSuite.SetupTest() // TRUNCATE + фикстуры
	s.copies = repositories.NewCopyRepository(s.db)
	s.loans = repositories.NewLoanRepository(s.db)
}

// reserveCopy — бронирует конкретный экземпляр, чтобы занять его в БД.
func (s *CopyRepositorySuite) reserveCopy(copyID uuid.UUID) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	l, err := loan.Reserve(copyID, fixtureReaderID, now)
	s.Require().NoError(err)
	s.Require().NoError(s.loans.Create(context.Background(), l))
}

// TestFindFree_Found — на чистых фикстурах оба экземпляра свободны,
// запрос обязан вернуть один из них. Какой именно — SQL с LIMIT 1
// не обещает, поэтому Contains, а не Equal.
func (s *CopyRepositorySuite) TestFindFree_Found() {
	got, err := s.copies.FindFreeCopyID(context.Background(), fixtureBookID)
	s.Require().NoError(err)
	s.Require().Contains([]uuid.UUID{fixtureCopyID, fixtureCopyID2}, got)
}

// TestFindFree_AllTaken — оба экземпляра заняты активными выдачами.
// Тест-кейс реального бага: до фикса скана uuid падал первый же вызов.
func (s *CopyRepositorySuite) TestFindFree_AllTaken() {
	s.reserveCopy(fixtureCopyID)
	s.reserveCopy(fixtureCopyID2)

	_, err := s.copies.FindFreeCopyID(context.Background(), fixtureBookID)
	s.Require().ErrorIs(err, ports.ErrNoFreeCopy)
}

// TestFindFree_SkipsDecommissioned — списанные экземпляры не выдаются,
// даже если на них нет активных выдач (ветка NOT c.decommissioned).
func (s *CopyRepositorySuite) TestFindFree_SkipsDecommissioned() {
	s.Require().NoError(s.db.Exec(`UPDATE copies SET decommissioned = true`).Error)

	_, err := s.copies.FindFreeCopyID(context.Background(), fixtureBookID)
	s.Require().ErrorIs(err, ports.ErrNoFreeCopy)
}

// TestFindFree_UnknownBook — несуществующая книга неотличима от книги
// без свободных экземпляров: та же ErrNoFreeCopy, а не паника или nil.
func (s *CopyRepositorySuite) TestFindFree_UnknownBook() {
	_, err := s.copies.FindFreeCopyID(context.Background(), uuid.New())
	s.Require().ErrorIs(err, ports.ErrNoFreeCopy)
}

func TestCopyRepositorySuite(t *testing.T) {
	suite.Run(t, new(CopyRepositorySuite))
}
