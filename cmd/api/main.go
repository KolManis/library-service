package main

import (
	"log"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/KolManis/library-service/internal/adapters/in/httphandlers"
	"github.com/KolManis/library-service/internal/adapters/out/repositories"
	"github.com/KolManis/library-service/internal/core/application/commands/issueloan"
	"github.com/KolManis/library-service/internal/core/application/commands/reservecopy"
	"github.com/KolManis/library-service/internal/core/application/commands/returnloan"
	"github.com/KolManis/library-service/internal/core/application/queries/getreaderloans"
	"github.com/KolManis/library-service/migrations"
)

func main() {
	// 1. Подключение к Postgres
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL не задан")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("ошибка подключения к БД: %v", err)
	}

	// 2. Миграции при старте — компромисс пет-проекта, в бою их гоняют
	// отдельным шагом деплоя до запуска приложения
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("получение sql.DB: %v", err)
	}
	// миграции вшиты в бинарник (go:embed) — работают из любой директории
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose: %v", err)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		log.Fatalf("миграции: %v", err)
	}

	// 3. Сборка зависимостей
	copyRepo := repositories.NewCopyRepository(db)
	loanRepo := repositories.NewLoanRepository(db)
	fineRepo := repositories.NewFineRepository(db)
	transactor := repositories.NewTransactor(db)

	reserveHandler := reservecopy.NewHandler(copyRepo, loanRepo, time.Now)
	issueHandler := issueloan.NewHandler(loanRepo, time.Now)
	returnHandler := returnloan.NewHandler(loanRepo, fineRepo, transactor, time.Now)
	readerLoansHandler := getreaderloans.NewHandler(loanRepo)

	httpReserveHandler := httphandlers.NewReserveHandler(reserveHandler)
	httpIssueHandler := httphandlers.NewIssueHandler(issueHandler)
	httpReturnHandler := httphandlers.NewReturnHandler(returnHandler)
	httpReaderLoansHandler := httphandlers.NewReaderLoansHandler(readerLoansHandler)

	// 4. Настройка Echo и маршрутов
	e := echo.New()
	e.POST("/books/:id/reserve", httpReserveHandler.Reserve)
	e.POST("/loans/:id/issue", httpIssueHandler.Issue)
	e.POST("/loans/:id/return", httpReturnHandler.Return)
	e.GET("/readers/:id/loans", httpReaderLoansHandler.List)

	// 5. Старт сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("сервер запущен на порту %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
