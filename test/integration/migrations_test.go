//go:build integration

package integration

import (
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
)

type MigrationsSuite struct {
	IntegrationSuite
}

// TestDownUp — обратимость миграций: Down-секции пишутся «на всякий
// случай» и без этого теста не выполняются никогда — ошибка в них
// всплыла бы только при откате на проде, в худший момент.
// Reset откатывает ВСЕ миграции до нуля, Up накатывает заново.
func (s *MigrationsSuite) TestDownUp() {
	sqlDB, err := s.db.DB()
	s.Require().NoError(err)

	s.Require().NoError(goose.Reset(sqlDB, "../../migrations"))
	s.Require().NoError(goose.Up(sqlDB, "../../migrations"))
}

func TestMigrationsSuite(t *testing.T) {
	suite.Run(t, new(MigrationsSuite))
}
