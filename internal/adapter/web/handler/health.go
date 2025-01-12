package handler

import (
	"github.com/tiagods/auth/internal/adapter/web/presenter/response"
	"github.com/tiagods/auth/internal/domain/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	health service.Health
}

func NewHealthHandler(health service.Health) HealthHandler {
	return HealthHandler{health: health}
}

func (h *HealthHandler) Health(c echo.Context) error {
	result := h.health.Check(c.Request().Context())
	responseResult := response.FromEntity(result)
	statusCode := http.StatusInternalServerError
	if responseResult.Status {
		statusCode = http.StatusOK
	}
	return c.JSON(statusCode, responseResult)
}
