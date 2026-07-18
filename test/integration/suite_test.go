//go:build integration

package integration

import (
	"context"

	"github.com/google/uuid"
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
}

// SetupSuite — один раз перед ВСЕМИ тестами сьюта:
// контейнер + миграции. Это ваш бывший setupDB почти дословно.
func (s *IntegrationSuite) SetupSuite() {
	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("library"),
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
}

// TearDownSuite — один раз после всех тестов: убить контейнер.
func (s *IntegrationSuite) TearDownSuite() {
	s.Require().NoError(testcontainers.TerminateContainer(s.container))
}

// SetupTest — перед КАЖДЫМ тестом: чистая БД.
// Дешёвая изоляция вместо нового контейнера.
func (s *IntegrationSuite) SetupTest() {
	s.Require().NoError(s.db.Exec(
		`TRUNCATE TABLE fines, loans, copies, readers, books RESTART IDENTITY CASCADE`,
	).Error)
}

// insertLoanFixtures создаёт книгу, экземпляр и читателя — родительские
// строки для loans (внешние ключи не дадут вставить выдачу без них).
func (s *IntegrationSuite) insertLoanFixtures(copyID, readerID uuid.UUID) {
	bookID := uuid.New()
	s.Require().NoError(s.db.Exec(
		`INSERT INTO books (id, title, author, isbn)
		 VALUES (?, 'Тестовая книга', 'Автор', '978-5-00000-000-0')`,
		bookID).Error)
	s.Require().NoError(s.db.Exec(
		`INSERT INTO copies (id, book_id) VALUES (?, ?)`,
		copyID, bookID).Error)
	s.Require().NoError(s.db.Exec(
		`INSERT INTO readers (id, full_name, email, status)
		 VALUES (?, 'Тестовый Читатель', 'reader@example.com', 'active')`,
		readerID).Error)
}
