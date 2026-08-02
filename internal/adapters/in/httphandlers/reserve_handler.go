package httphandlers

import (
	"errors"
	"net/http"

	"github.com/KolManis/library-service/internal/core/application/commands/reservecopy"
	"github.com/KolManis/library-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ReserveHandler — переводчик HTTP → команда reservecopy → HTTP.
type ReserveHandler struct {
	reserve *reservecopy.Handler
}

// NewReserveHandler создаёт ReserveHandler поверх готового command-хендлера.
func NewReserveHandler(reserve *reservecopy.Handler) *ReserveHandler {
	return &ReserveHandler{reserve: reserve}
}

// ReserveRequest — тело запроса POST /books/:id/reserve.
type ReserveRequest struct {
	ReaderID uuid.UUID `json:"reader_id"`
}

// Reserve обрабатывает POST /books/:id/reserve: бронирует свободный экземпляр
// книги за читателем. 201 с loan_id при успехе, 409 — свободных экземпляров нет.
func (h *ReserveHandler) Reserve(c echo.Context) error {
	// 1. Распарсить bookID из URL
	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "невалидный bookID",
		})
	}

	// 2. Распарсить JSON-тело
	var req ReserveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "невалидное тело запроса",
		})
	}

	// 3. Создать команду (проверка ID внутри)
	cmd, err := reservecopy.NewCommand(bookID, req.ReaderID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 4. Выполнить сценарий
	loanID, err := h.reserve.Handle(c.Request().Context(), cmd)
	if err != nil {
		// Маппинг доменных ошибок в HTTP-коды
		if errors.Is(err, ports.ErrNoFreeCopy) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "нет свободных экземпляров",
			})
		}
		if errors.Is(err, ports.ErrConcurrentReservation) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "экземпляр только что заняли, попробуйте ещё раз",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "внутренняя ошибка сервера",
		})
	}

	// 5. Успех — вернуть 201 с loan_id
	return c.JSON(http.StatusCreated, map[string]string{
		"loan_id": loanID.String(),
	})
}
