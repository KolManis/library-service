package httphandlers

import (
	"net/http"
	"time"

	"github.com/KolManis/library-service/internal/core/application/queries/getreaderloans"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// readerLoanDTO — JSON-представление одной выдачи в ответе GET /readers/:id/loans.
type readerLoanDTO struct {
	LoanID     uuid.UUID  `json:"loan_id"`
	CopyID     uuid.UUID  `json:"copy_id"`
	Status     string     `json:"status"`
	ReservedAt time.Time  `json:"reserved_at"`
	IssuedAt   *time.Time `json:"issued_at,omitempty"`
	DueAt      *time.Time `json:"due_at,omitempty"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`
}

// ReaderLoansHandler — переводчик HTTP → запрос getreaderloans → HTTP.
type ReaderLoansHandler struct {
	query *getreaderloans.Handler
}

// NewReaderLoansHandler создаёт ReaderLoansHandler поверх готового
// query-хендлера.
func NewReaderLoansHandler(query *getreaderloans.Handler) *ReaderLoansHandler {
	return &ReaderLoansHandler{query: query}
}

// List обрабатывает GET /readers/:id/loans: возвращает список выдач читателя
// (включая завершённые). Пустой список — 200 с пустым JSON-массивом, не ошибка.
func (h *ReaderLoansHandler) List(c echo.Context) error {
	readerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return respondBadRequest(c, "невалидный readerID")
	}

	q, err := getreaderloans.NewQuery(readerID)
	if err != nil {
		return respondBadRequest(c, err.Error())
	}

	views, err := h.query.Handle(c.Request().Context(), q)
	if err != nil {
		return respondError(c, err)
	}

	dtos := make([]readerLoanDTO, 0, len(views))
	for _, v := range views {
		dtos = append(dtos, readerLoanDTO{
			LoanID:     v.LoanID,
			CopyID:     v.CopyID,
			Status:     v.Status,
			ReservedAt: v.ReservedAt,
			IssuedAt:   v.IssuedAt,
			DueAt:      v.DueAt,
			ReturnedAt: v.ReturnedAt,
		})
	}

	return c.JSON(http.StatusOK, dtos)
}
