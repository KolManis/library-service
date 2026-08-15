package httphandlers

import (
	"net/http"

	"github.com/KolManis/library-service/internal/core/application/commands/reservecopy"
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
		return respondBadRequest(c, "невалидный bookID")
	}

	// 2. Распарсить JSON-тело
	var req ReserveRequest
	if err := c.Bind(&req); err != nil {
		return respondBadRequest(c, "невалидное тело запроса")
	}

	// 3. Создать команду (проверка ID внутри)
	cmd, err := reservecopy.NewCommand(bookID, req.ReaderID)
	if err != nil {
		return respondBadRequest(c, err.Error())
	}

	// 4. Выполнить сценарий
	loanID, err := h.reserve.Handle(c.Request().Context(), cmd)
	if err != nil {
		return respondError(c, err)
	}

	// 5. Успех — вернуть 201 с loan_id
	return c.JSON(http.StatusCreated, map[string]string{
		"loan_id": loanID.String(),
	})
}
