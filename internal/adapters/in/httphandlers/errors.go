package httphandlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/KolManis/library-service/internal/core/domain/aggregates/loan"
	"github.com/KolManis/library-service/internal/core/ports"
)

// errorMapping связывает доменную/инфраструктурную ошибку с HTTP-кодом и текстом.
var errorMapping = []struct {
	err  error
	code int
	msg  string
}{
	{ports.ErrNoFreeCopy, http.StatusConflict, "нет свободных экземпляров"},
	{ports.ErrConcurrentReservation, http.StatusConflict, "экземпляр только что заняли, попробуйте ещё раз"},
	{ports.ErrLoanNotFound, http.StatusNotFound, "выдача не найдена"},
	{loan.ErrInvalidTransition, http.StatusConflict, "недопустимый переход статуса"},
}

// respondError маппит ошибку на HTTP-ответ по таблице выше. Не нашли — 500.
func respondError(c echo.Context, err error) error {
	for _, m := range errorMapping {
		if errors.Is(err, m.err) {
			return c.JSON(m.code, map[string]string{"error": m.msg})
		}
	}
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "внутренняя ошибка сервера"})
}

// respondBadRequest — 400 с заданным текстом (ошибка валидации запроса).
func respondBadRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": msg})
}
