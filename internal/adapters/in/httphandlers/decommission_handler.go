package httphandlers

import (
	"net/http"

	"github.com/KolManis/library-service/internal/core/application/commands/decommissioncopy"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DecommissionHandler — переводчик HTTP → команда decommissioncopy → HTTP.
type DecommissionHandler struct {
	decommission *decommissioncopy.Handler
}

// NewDecommissionHandler создаёт DecommissionHandler поверх готового command-хендлера.
func NewDecommissionHandler(decommission *decommissioncopy.Handler) *DecommissionHandler {
	return &DecommissionHandler{decommission: decommission}
}

// Decommission обрабатывает POST /copies/:id/decommission: списывает экземпляр,
// отменяет его reserved-бронь если есть. 200 при успехе, 404 — экземпляр не найден.
func (h *DecommissionHandler) Decommission(c echo.Context) error {
	copyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return respondBadRequest(c, "невалидный copyID")
	}

	cmd, err := decommissioncopy.NewCommand(copyID)
	if err != nil {
		return respondBadRequest(c, err.Error())
	}

	if err := h.decommission.Handle(c.Request().Context(), cmd); err != nil {
		return respondError(c, err)
	}
	return c.NoContent(http.StatusOK)
}
