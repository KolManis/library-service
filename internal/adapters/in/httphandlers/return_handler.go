package httphandlers

import (
	"net/http"

	"github.com/KolManis/library-service/internal/core/application/commands/returnloan"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ReturnHandler — переводчик HTTP → команда returnloan → HTTP.
type ReturnHandler struct {
	returnLoan *returnloan.Handler
}

// NewReturnHandler создаёт ReturnHandler поверх готового command-хендлера.
func NewReturnHandler(returnLoan *returnloan.Handler) *ReturnHandler {
	return &ReturnHandler{returnLoan: returnLoan}
}

// Return обрабатывает POST /loans/:id/return: возвращает книгу, при возврате
// из overdue начисляет штраф. 200 при успехе, 404 — выдача не найдена, 409 —
// недопустимый переход статуса.
func (h *ReturnHandler) Return(c echo.Context) error {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return respondBadRequest(c, "невалидный loanID")
	}

	cmd, err := returnloan.NewCommand(loanID)
	if err != nil {
		return respondBadRequest(c, err.Error())
	}

	err = h.returnLoan.Handle(c.Request().Context(), cmd)
	if err != nil {
		return respondError(c, err)
	}
	return c.NoContent(http.StatusOK)
}
