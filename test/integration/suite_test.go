//go:build integration

package integration

import (
	"context"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// IntegrationSuite — общая база: контейнер, миграции, изоляция.
// Конкретные сьюты встраивают её и получают всё бесплатно.
type IntegrationSuite struct {
	suite.Suite
	db        *gorm.DB
	container *tcpostgres.PostgresContainer
	fixtures  *testfixtures.Loader
}

// SetupSuite — один раз перед ВСЕМИ тестами сьюта:
// контейнер + миграции.
func (s *IntegrationSuite) SetupSuite() {
	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("library_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	s.Require().NoError(err)
	s.container = container

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	s.db, err = gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	s.Require().NoError(err)

	sqlDB, err := s.db.DB()
	s.Require().NoError(err)
	s.Require().NoError(goose.SetDialect("postgres"))
	s.Require().NoError(goose.Up(sqlDB, "../../migrations"))

	s.fixtures, err = testfixtures.New(
		testfixtures.Database(sqlDB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures"),
	)
	s.Require().NoError(err)
}

// TearDownSuite — один раз после всех тестов: убить контейнер.
func (s *IntegrationSuite) TearDownSuite() {
	s.Require().NoError(testcontainers.TerminateContainer(s.container))
}

// SetupTest — перед КАЖДЫМ тестом: чистая БД.
func (s *IntegrationSuite) SetupTest() {
	s.Require().NoError(s.db.Exec(`TRUNCATE TABLE fines, loans, copies, readers, books RESTART IDENTITY CASCADE`).Error)
	s.Require().NoError(s.fixtures.Load())
}
