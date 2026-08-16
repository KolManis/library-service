package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"

	"github.com/KolManis/library-service/internal/adapters/in/httphandlers"
	"github.com/KolManis/library-service/internal/adapters/out/repositories"
	"github.com/KolManis/library-service/internal/config"
	"github.com/KolManis/library-service/internal/core/application/commands/issueloan"
	"github.com/KolManis/library-service/internal/core/application/commands/reservecopy"
	"github.com/KolManis/library-service/internal/core/application/commands/returnloan"
	"github.com/KolManis/library-service/internal/core/application/queries/getreaderloans"
	"github.com/KolManis/library-service/internal/infra"
	"github.com/KolManis/library-service/migrations"
)

func main() {
	infra.NewLogger()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("ошибка конфигурации", "error", err)
		os.Exit(1)
	}

	db, err := infra.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		slog.Error("подключение к БД", "error", err)
		os.Exit(1)
	}

	// 2. Миграции при старте — компромисс пет-проекта, в бою их гоняют
	// отдельным шагом деплоя до запуска приложения
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

	ctx, stop := infra.ShutdownContext()
	defer stop()

	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("ошибка сервера", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("сервер запущен", "port", cfg.Port)

	<-ctx.Done()
	slog.Info("получен сигнал остановки, завершаем работу")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("ошибка graceful shutdown", "error", err)
	}
}
