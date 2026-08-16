package infra

import (
	"log/slog"
	"os"
)

// NewLogger настраивает JSON-логгер и устанавливает его как глобальный.
func NewLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	return logger
}
