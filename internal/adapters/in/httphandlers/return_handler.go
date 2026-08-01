package httphandlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/KolManis/library-service/internal/core/application/commands/returnloan"
	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
)

type ReturnHandler struct {
	returnLoan *returnloan.Handler
}

func NewReturnHandler(returnLoan *returnloan.Handler) *ReturnHandler {
	return &ReturnHandler{returnLoan: returnLoan}
}

func (h *ReturnHandler) Return(c echo.Context) error {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "невалидный loanID",
		})
	}

	cmd, err := returnloan.NewCommand(loanID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	err = h.returnLoan.Handle(c.Request().Context(), cmd)
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
