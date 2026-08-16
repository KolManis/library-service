package infra

import (
	"context"
	"os/signal"
	"syscall"
)

// ShutdownContext возвращает контекст, отменяемый при SIGINT/SIGTERM.
func ShutdownContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
}
