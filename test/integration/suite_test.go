//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

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
	db       *gorm.DB
	fixtures *testfixtures.Loader
}

var sharedDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("library_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		panic(err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	sharedDB, err = gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	sqlDB, err := sharedDB.DB()
	if err != nil {
		panic(err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}
	if err := goose.Up(sqlDB, "../../migrations"); err != nil {
		panic(err)
	}

	code := m.Run()

	if err := testcontainers.TerminateContainer(container); err != nil {
		panic(err)
	}
	os.Exit(code)
}

// SetupSuite — один раз перед тестами КАЖДОЙ сьюты: взять общую БД, настроить фикстуры.
func (s *IntegrationSuite) SetupSuite() {
	s.db = sharedDB

	sqlDB, err := s.db.DB()
	s.Require().NoError(err)

	s.fixtures, err = testfixtures.New(
		testfixtures.Database(sqlDB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("fixtures"),
	)
	s.Require().NoError(err)
}

// SetupTest — перед КАЖДЫМ тестом: чистая БД.
func (s *IntegrationSuite) SetupTest() {
	s.Require().NoError(s.db.Exec(`TRUNCATE TABLE fines, loans, copies, readers, books, outbox RESTART IDENTITY CASCADE`).Error)
	s.Require().NoError(s.fixtures.Load())
}
