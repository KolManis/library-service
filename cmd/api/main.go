package main

import (
	"log/slog"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL не задан")
		os.Exit(1)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("ошибка подключения к БД", "error", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("получение sql.DB", "error", err)
		os.Exit(1)
	}
	// миграции вшиты в бинарник (go:embed) — работают из любой директории
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("goose", "error", err)
		os.Exit(1)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		slog.Error("миграции", "error", err)
		os.Exit(1)
	}

	copyRepo := repositories.NewCopyRepository(db)
	loanRepo := repositories.NewLoanRepository(db)
	fineRepo := repositories.NewFineRepository(db)
	outboxRepo := repositories.NewOutboxRepository(db)
	transactor := repositories.NewTransactor(db)

	reserveHandler := reservecopy.NewHandler(copyRepo, loanRepo, outboxRepo, transactor, time.Now)
	issueHandler := issueloan.NewHandler(loanRepo, outboxRepo, transactor, time.Now)
	returnHandler := returnloan.NewHandler(loanRepo, fineRepo, outboxRepo, transactor, time.Now)
	readerLoansHandler := getreaderloans.NewHandler(loanRepo)

	httpReserveHandler := httphandlers.NewReserveHandler(reserveHandler)
	httpIssueHandler := httphandlers.NewIssueHandler(issueHandler)
	httpReturnHandler := httphandlers.NewReturnHandler(returnHandler)
	httpReaderLoansHandler := httphandlers.NewReaderLoansHandler(readerLoansHandler)

	e := echo.New()
	e.POST("/books/:id/reserve", httpReserveHandler.Reserve)
	e.POST("/loans/:id/issue", httpIssueHandler.Issue)
	e.POST("/loans/:id/return", httpReturnHandler.Return)
	e.GET("/readers/:id/loans", httpReaderLoansHandler.List)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("сервер запущен", "port", port)
	e.Logger.Fatal(e.Start(":" + port))
}
