package httphandlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/KolManis/library-service/internal/core/application/commands/issueloan"
	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
)

// IssueHandler — переводчик HTTP → команда → HTTP.
type IssueHandler struct {
	issue *issueloan.Handler
}

func NewIssueHandler(issue *issueloan.Handler) *IssueHandler {
	return &IssueHandler{issue: issue}
}

// Issue обрабатывает POST /loans/:id/issue.
func (h *IssueHandler) Issue(c echo.Context) error {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "невалидный loanID"},
		)
	}

	cmd, err := issueloan.NewCommand(loanID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	err = h.issue.Handle(c.Request().Context(), cmd)
	if err != nil {
		if errors.Is(err, ports.ErrLoanNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "выдача не найдена",
			})
		}
		if errors.Is(err, loan.ErrInvalidTransition) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "недопустимый переход статуса",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "внутренняя ошибка сервера",
		})
	}

	return c.NoContent(http.StatusOK)
}
