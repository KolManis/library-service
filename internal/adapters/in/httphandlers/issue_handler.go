package httphandlers

import (
	"net/http"

	"github.com/KolManis/library-service/internal/core/application/commands/issueloan"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// IssueHandler — переводчик HTTP → команда issueloan → HTTP.
type IssueHandler struct {
	issue *issueloan.Handler
}

// NewIssueHandler создаёт IssueHandler поверх готового command-хендлера.
func NewIssueHandler(issue *issueloan.Handler) *IssueHandler {
	return &IssueHandler{issue: issue}
}

// Issue обрабатывает POST /loans/:id/issue: выдаёт книгу читателю на руки.
// 200 при успехе, 404 — выдача не найдена, 409 — недопустимый переход статуса.
func (h *IssueHandler) Issue(c echo.Context) error {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return respondBadRequest(c, "невалидный loanID")
	}

	cmd, err := issueloan.NewCommand(loanID)
	if err != nil {
		return respondBadRequest(c, err.Error())
	}

	err = h.issue.Handle(c.Request().Context(), cmd)
	if err != nil {
		return respondError(c, err)
	}

	return c.NoContent(http.StatusOK)
}
